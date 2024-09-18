trigger_mode(TRIGGER_MODE_MANUAL)
allow_k8s_contexts('orbstack')

load('ext://namespace', 'namespace_create')


docker_build(
  'localhost:5001/magiccicd/app-of-app-visualizer',
  '.'
)

namespace_create('app-of-app-visualizer-system-tilt')
k8s_resource(
  new_name = 'namespace',
  objects = ['app-of-app-visualizer-system-tilt:namespace'],
  labels = ['app-of-app-visualizer']
)

k8s_yaml(
  helm(
    './charts/app-of-apps-visualizer',
    name = 'app-of-apps-visualizer',
    namespace = 'app-of-app-visualizer-system-tilt',
    # values = 'hack/tilt/values.dev.yaml'
  )
)

k8s_resource(
  new_name = 'crds',
  objects = [
    'appofapps.visualization.magiccicd:customresourcedefinition',
    'appversions.visualization.magiccicd:customresourcedefinition'
  ],
  labels = ['app-of-app-visualizer']
)