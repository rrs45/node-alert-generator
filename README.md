# Node alert generator
This is a K8s custom controller to monitor Node health as represented by node-problem-detector. It does the following:
1. Extracts node conditions based on following conditions: 
   * Name starts with 'NPD-'
   * Status is 'True'
   * Node does not have maintenance labels except the ones created by node-alert-generator
   * Alerts not older than given time period
2. Populates given config map with the alerts by:
   * Gathering all filtered conditions upto the user defined interval
   * Deduplicating alerts
   * Populating the given config map with alerts if they are different than alerts from previous interval
   * Labeling the Node with 'maintenance.box.com/source=npd' if enabled by user

## Flow
![Autoremediation_Image](k8s_automediation.jpeg)

## Usage
```$ ./node-alert-generator -h
Usage of ./node-alert-generator:
  -address string
    	Address to bind the alert generator server. (default "127.0.0.1")
  -alert-config-map string
    	Name of config map to store alerts (default "npd-alerts")
  -apiserver-host  based string
    	Custom hostname used to connect to Kubernetes ApiServer
  -interval string
    	Interval in seconds at which configmap will be updated (default "60")
  -log-file string
    	Log file to store all logs (default "/var/log/service/node-alert-generator.log")
  -namespace string
    	Namespace where config map will be (default "node-alert-generator")
  -nolabel
    	Dont set labels (default true)
  -port string
    	Port to bind the alert generator server for /healthz endpoint (default "8080")
...<snipping irrelevant options>...
```
