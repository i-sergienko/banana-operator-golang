1) Install Go https://golang.org/doc/install (using go 1.15.5 here)
2) Install Operator SDK CLI https://sdk.operatorframework.io/docs/installation/ (using v1.3.0-36-gfbce7e7b here)

3) Create and enter the project directory:  
   `mkdir banana-operator-golang`  
   `cd banana-operator-golang`

4) Initialize the project using Operator SDK CLI:  
   `operator-sdk init --domain fruits.com --repo github.com/i-sergienko/banana-operator-golang`  
   `--domain` - domain for resource groups. Group from `spec.group` from the CRD will end in this domain.  
   `--repo` - this is just the module name - doesn't have to be a URL, or reference a GitHub repo.  
   Notable generated output:  
   *go.mod* - Go module file  
   *main.go* - the main class of the application.


5) Create an API - the Custom Resource + Custom Controller
   `operator-sdk create api --version v1 --kind Banana --resource --controller`
   `--group` - the subdomain part for `spec.group`. It's appended to `--domain` you specified earlier - if
   your `spec.group` is the same as domain (as in this example), don't specify this parameter.  
   `--version` - the version name from CRD  
   `--kind` - the resource kind from CRD
   `--resource` - generate Go model classes for the new resource
   `--controller` - generate the controller class template Notable generated output:  
   *api/v1/banana_types.go* - model classes. You'll modify them manually later.
   *config/rbac/banana_editor_role.yaml*, *config/rbac/banana_viewer_role.yaml* - cluster roles for Banana resource  
   *controllers/banana_controller.go* - the controller class with event handling/reconciliation logic
   *suite_test.go* - the test suite for your new controller


6) Add fields to the `BananaSpec`/`BananaStatus` model classes in *api/v1/banana_types.go*.  
   You normally don't need to touch anything else there. Run `make generate` every time you modify `BananaSpec`
   /`BananaStatus` - it will re-generate *api/v1/zz_generated.deepcopy.go* (you shouldn't touch this file manually, but
   it's necessary for the app to function).


7) Generate CRD manifests by running:  
   `make manifests`  
   This will generate:
    * The CustomResourceDefinition, with the fields/validation already defined, in *
      config/crd/bases/fruits.com_bananas.yaml*. You might want to configure additional validation rules for `spec`
      /`status` here.
    * A `manager-role` ClusterRole with permissions to do anything to `Banana` resources, in *config/rbac/role.yaml*.
      You'll likely want to rename the generated role before deploying the app, if there are multiple Operators running
      in the cluster (which is typical).

8) Implement the event-handling (reconciliation) logic in `BananaReconciler.Reconcile` method (in *
   controllers/banana_controller.go* file).  
   Also place annotations here for ClusterRole permissions for all the needed resources, like
   this `// +kubebuilder:rbac:groups=fruits.com,resources=bananas/finalizers,verbs=update`.  
   Run `make manifests` again if you modified the permissions.

9) Implement tests in *controllers/suite_test.go*. https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html  
   Launch a Kubernetes cluster (in a CI environment it's easy to use kind), or use an existing one.  
   Remove the `test` dependency from `docker-build: test` target in Makefile.  
   Remove the Envtest setup from `test` target: we're using an existing cluster, so just leave `go test ./... -coverprofile cover.out`.  
     
   Set the `USE_EXISTING_CLUSTER=true` environment variable - that way the tests will use your existing `$HOME/.kube/config` - and run `make test`.  
   NOTE FOR WINDOWS USERS: if you develop on Windows, but run tests in a Linux CI pipeline, check the `generate` target
   in your *Makefile* - there shouldn't be any backslashes (`\\`) - if there are, replace each pair with a forward
   slash (`/`), like this:  
   Before: `$(CONTROLLER_GEN) object:headerFile="hack\\boilerplate.go.txt" paths="./..."`  
   After: `$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."`