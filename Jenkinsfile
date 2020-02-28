@Library("jenkins-pipeline-library") _

pipeline {
    agent { label 'skynet' }
    options {
        timeout(time: 10, unit: 'HOURS')
    }
    environment {
        SKYNET_APP = 'node-alert-generator'
    }
    parameters {
        string(name: "BUILD_NUMBER", defaultValue: "", description: "Replay build value")
    }
    stages {
        stage('Build') {
            when { branch 'master'  }
            steps {
                githubCheck(
                    'Build Image': {
                        buildImage()
                        echo "Just built image with id ${builtImage.imageId}"
                    }
                )
            }
        }
       stage('Deploy To Sandbox') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'sandbox', app: SKYNET_APP, watch: false, canary: false
            }
        } 
       stage('Deploy To DSV31') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'dsv31', app: SKYNET_APP, watch: false, canary: false
            }
        } 
      stage('Deploy To DSV31-K8S-C1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'dsv31-k8s-c1', app: SKYNET_APP, watch: false, canary: false
            }
        } 
       stage('Deploy To VSV1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'vsv1', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To VSV1-K8S-C1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'vsv1-k8s-c1', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To LV7') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'lv7', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To US-RNO-A-K8S-C1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'us-rno-a-k8s-c1', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To US-RNO-A-K8S-C2') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'us-rno-a-k8s-c2', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To LV7-K8S-C1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'lv7-k8s-c1', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To US-LAS-B-K8S-C1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'us-las-b-k8s-c1', app: SKYNET_APP, watch: false, canary: false
            }
        }
        
        
    }
        post {
        always {
            archiveBuildInfo()
        }
    }
}
