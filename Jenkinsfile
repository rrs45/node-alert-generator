@Library("jenkins-pipeline-library") _

pipeline {
    agent { label 'skynet' }
    options {
        timeout(time: 1, unit: 'HOURS')
    }
    environment {
        SKYNET_APP = 'node-alert-generator'
    }
    parameters {
        string(name: "BUILD_NUMBER", defaultValue: "", description: "Replay build value")
    }
    stages {
        stage('Build') {
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
           // when { branch 'master'  }
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
       stage('Deploy To VSV1') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'dsv31', app: SKYNET_APP, watch: false, canary: false
            }
        }
        stage('Deploy To LV7') {
            when { branch 'master'  }
            steps {
                deploy cluster: 'dsv31', app: SKYNET_APP, watch: false, canary: false
            }
        }

    }
        post {
        always {
            archiveBuildInfo()
        }
    }
}
