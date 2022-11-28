# Copyright (c) 2022 RaptorML authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import inspect
import types as pytypes
from typing import Union, List

from . import replay, local_state, config
from .program import Program
from .types import FeatureSpec, AggrSpec, ResourceReference, AggrFn, BuilderSpec, FeatureSetSpec
from .program import normalize_fqn

def _wrap_decorator_err(f):
    def wrap(*args, **kwargs):
        try:
            return f(*args, **kwargs)
        except Exception as e:
            back_frame = e.__traceback__.tb_frame.f_back
            tb = pytypes.TracebackType(tb_next=None,
                                       tb_frame=back_frame,
                                       tb_lasti=back_frame.f_lasti,
                                       tb_lineno=back_frame.f_lineno)
            raise Exception(f"in {args[0].__name__}: {str(e)}").with_traceback(tb)

    return wrap


@_wrap_decorator_err
def _opts(func, options: dict):
    if hasattr(func, "raptor_spec"):
        raise Exception("option decorators must be before the register decorator")
    if not hasattr(func, "__raptor_options"):
        func.__raptor_options = {}

    for k, v in options.items():
        func.__raptor_options[k] = v
    return func


def aggr(funcs: [AggrFn], granularity=None):
    """
    Register aggregations for the Feature Definition.
    :param granularity: the granularity of the aggregation (this is overriding the freshness).
    :param funcs: a list of :func:`AggrFn`
    """

    def decorator(func):
        for f in funcs:
            if f == AggrFn.Unknown:
                raise Exception("Unknown aggr function")
        return _opts(func, {"aggr": AggrSpec(funcs, granularity)})

    return decorator


def data_source(name: str, namespace: str = None):
    """
    Register a DataSource for the FeatureDefinition.
    :param name: the name of the DataSource.
    :param namespace: the namespace of the DataSource.
    """

    def decorator(func):
        return _opts(func, {"data_source": ResourceReference(name, namespace)})

    return decorator


def namespace(namespace: str):
    """
    Register a namespace for the Feature Definition.
    :param namespace: namespace name
    """

    def decorator(func):
        return _opts(func, {"namespace": namespace})

    return decorator


def builder(kind: str, options=None):
    """
    Register a builder for the Feature Definition.
    :param kind: the kind of the builder.
    :param options: options for the builder.
    :return:
    """

    if options is None:
        options = {}

    def decorator(func):
        return _opts(func, {"builder": BuilderSpec(kind, options)})

    return decorator


def register(keys: Union[str, List[str]], staleness: str, freshness: str = '', options=None):
    """
    Register a Feature Definition within the LabSDK.

    A feature definition is a PyExp handler function that process a calculation request and calculates
    a feature value for it::
        :param RaptorRequest **kwargs: a request bag dictionary(:class:`RaptorRequest`)
        :return: a tuple of (value, timestamp, entity_id) where:
            - value: the feature value
            - timestamp: the timestamp of the value - when None, it uses the request timestamp.
            - entity_id: the entity id of the value - when None, it uses the request entity_id.
                If the request entity_id is None, it's required to specify an entity_id

    :Example:
        @register(primitive="my_primitive", freshness="1h", staleness="1h")
        def my_feature(**req):
            return "Hello "+ req["entity_id"]

    :param staleness: the staleness of the feature.
    :param freshness: the freshness of the feature.
    :param options: optional options for the feature.
    :return: a registered Feature Definition
    """
    if options is None:
        options = {}

    if isinstance(keys, str):
        keys = [keys]

    @_wrap_decorator_err
    def decorator(func):
        spec = FeatureSpec()
        spec.freshness = freshness
        spec.staleness = staleness
        spec.keys = keys
        spec.name = func.__name__
        spec.description = func.__doc__

        if hasattr(func, "__raptor_options"):
            for k, v in func.__raptor_options.items():
                options[k] = v

        # append annotations
        if "builder" in options:
            spec.builder = options['builder']

        if "namespace" in options:
            spec.namespace = options['namespace']

        if "data_source" in options:
            spec.data_source = options['data_source']
        if "aggr" in options:
            for f in options['aggr'].funcs:
                if not f.supports(spec.primitive):
                    raise Exception(
                        f"{func.__name__} aggr function {f} not supported for primitive {spec.primitive}")
            spec.aggr = options['aggr']

        if freshness == '' and (spec.aggr is None or spec.aggr.granularity is None):
            raise Exception(f"{func.__name__} must have a freshness or an aggregation with granularity")
        if staleness == '':
            raise Exception(f"{func.__name__} must have a staleness")

        # parse the program
        def fqn_resolver(obj):
            frame = inspect.currentframe().f_back.f_back

            feat: Union[FeatureSpec, None] = None
            if obj in frame.f_globals:
                if hasattr(frame.f_globals[obj], "raptor_spec"):
                    feat = frame.f_globals[obj].raptor_spec
            elif obj in frame.f_locals:
                if hasattr(frame.f_locals[obj], "raptor_spec"):
                    feat = frame.f_locals[obj].raptor_spec
            if feat is None:
                raise Exception(f"Cannot resolve {obj} to an FQN")

            if feat.aggr is not None:
                raise Exception("You must specify a FQN with AggrFn(i.e. `namespace.name+sum`) for aggregated features")

            return feat.fqn()

        spec.program = Program(func, fqn_resolver)
        spec.primitive = spec.program.primitive

        # register
        func.raptor_spec = spec
        func.replay = replay.new_replay(spec)
        func.manifest = spec.manifest
        func.export = spec.manifest
        local_state.register_spec(spec)

        if hasattr(func, "__raptor_options"):
            del func.__raptor_options

        return func

    return decorator


def feature_set(register=False, options=None):
    """
    Register a FeatureSet Definition.

    A feature set definition in the LabSDK is constituted by a function that returns a list of features.
        You can specify a feature in the list using its FQN or via a variable that hold's the feature definition.
        When specifying a feature that have aggregations, you **must** specify the feature using its FQN.

    :Example:
        @feature_set(register=True)
        def my_feature_set():
            return [my_feature, "my_other_feature.default[sum]"]


    :param register: if True, the feature set will be registered in the LabSDK, and you'll be able to export
        it's manifest.
    :param options:
    :return:
    """
    if options is None:
        options = {}

    @_wrap_decorator_err
    def decorator(func):
        if inspect.signature(func) != inspect.signature(lambda: []):
            raise Exception(f"{func.__name__} have an invalid signature for a FeatureSet definition")

        fts = []
        for f in func():
            if type(f) is str:
                local_state.spec_by_fqn(f)
                fts.append(normalize_fqn(f, config.default_namespace))
            if callable(f):
                ft = local_state.spec_by_src_name(f.__name__)
                if ft is None:
                    raise Exception("Feature not found")
                if ft.aggr is not None:
                    raise Exception(
                        "You must specify a FQN with AggrFn(i.e. `namespace.name+sum`) for aggregated features")
                fts.append(ft.fqn())

        if hasattr(func, "__raptor_options"):
            for k, v in func.__raptor_options.items:
                options[k] = v

        spec = FeatureSetSpec()
        spec.name = func.__name__
        spec.description = func.__doc__
        spec.features = fts

        if "key_feature" in options:
            spec.key_feature = options["key_feature"]

        if "namespace" in options:
            spec.namespace = options["namespace"]

        if "timeout" not in options:
            options["timeout"] = "5s"

        spec.timeout = options["timeout"]
        spec.name = func.__name__
        spec.description = func.__doc__

        func.raptor_spec = spec
        func.historical_get = replay.new_historical_get(spec)
        func.manifest = spec.manifest
        func.export = spec.manifest
        local_state.register_spec(spec)

        if hasattr(func, "__raptor_options"):
            del func.__raptor_options

        if register:
            local_state.register_spec(spec)
        return func

    return decorator
