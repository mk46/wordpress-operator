Proof of Concept (POC) Requirement: Kubernetes Operator Development
 
Objective
Develop a Kubernetes Operator to manage a custom web application Wordpress. The Operator should automate the deployment, configuration, and scaling of the Wordpress application, as well as manage the associated resources such as ConfigMaps, Secrets, and Horizontal Pod Autoscalers (HPA).
 
Scope
- Create a Custom Resource Definition (CRD) for the Wordpress application.
- Develop a controller to manage the lifecycle of the custom resource.
- Handle the creation and update of ConfigMaps, Secrets, Deployments, and HPA.
- Ensure the Operator can scale the application based on CPU usage.
 
Deliverables
- CRD YAML file for the custom web application.
- Kubernetes Operator implementation using Operator SDK (in Go).
- Documentation on how to deploy and use the Operator.
 
Technical Requirements
 
1. Custom Resource Definition (CRD)
   - Define a CRD named `Wordpress`.
   - CRD should include the following specifications:
     - `image`: Docker image of the wordpress application.
     - `replicas`: Number of replicas for the Deployment.
     - `configData`: Configuration data to be stored in a ConfigMap.
     - `dbUsername` and `dbPassword`: Database credentials to be stored in a Secret.
     - `minReplicas` and `maxReplicas`: Minimum and maximum replicas for HPA.
     - `targetCPUUtilizationPercentage`: Target CPU utilization for HPA.
 
2. Controller
   - Implement the controller to watch for changes in `Wordpress` custom resources.
   - Reconcile the desired state by creating or updating the following Kubernetes resources:
     - ConfigMap: Store configuration data.
     - Secret: Store database credentials.
     - Deployment: Deploy the Wordpress application.
     - Horizontal Pod Autoscaler (HPA): Scale the application based on CPU usage.
 
3. Reconciliation Logic
   - Fetch the `Wordpress` custom resource.
   - Create or update the ConfigMap with `configData`.
   - Create or update the Secret with `dbUsername` and `dbPassword`.
   - Create or update the Deployment with the specified `image` and `replicas`.
   - Create or update the HPA with the specified `minReplicas`, `maxReplicas`, and `targetCPUUtilizationPercentage`.
 
4. Testing
   - Deploy the CRD and Operator in a Kubernetes cluster.
   - Create a sample `Wordpress` custom resource and verify that all associated resources (ConfigMap, Secret, Deployment, HPA) are created and updated correctly.
   - Test scaling by adjusting the CPU load and observing the HPA behavior.
 
Documentation
Provide detailed instructions on how to deploy and use the Operator, including:
 
- Prerequisites (e.g., Kubernetes cluster, kubectl, Operator SDK).
- Steps to build and deploy the Operator.
- Example of creating a `Wordpress` custom resource.
- How to observe and verify the Operator's actions (e.g., checking ConfigMap, Secret, Deployment, and HPA).