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

    }
        post {
        always {
            archiveBuildInfo()
        }
    }
}
