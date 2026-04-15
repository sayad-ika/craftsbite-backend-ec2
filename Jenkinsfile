pipeline {
    agent any

    environment {
        AWS_DEFAULT_REGION = 'ap-southeast-1'
        AWS_CREDENTIALS_ID = 'aws-craftsbite-credentials'
    }

    options {
        disableConcurrentBuilds()
        timeout(time: 15, unit: 'MINUTES')
    }

    stages {
        stage('Init') {
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId   : "${AWS_CREDENTIALS_ID}",
                    accessKeyVariable: 'AWS_ACCESS_KEY_ID',
                    secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'
                ]]) {
                    sh 'terraform -chdir=terraform init -input=false'
                }
            }
        }

        stage('Validate') {
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId   : "${AWS_CREDENTIALS_ID}",
                    accessKeyVariable: 'AWS_ACCESS_KEY_ID',
                    secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'
                ]]) {
                    sh 'terraform -chdir=terraform validate'
                }
            }
        }

        stage('Plan') {
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId   : "${AWS_CREDENTIALS_ID}",
                    accessKeyVariable: 'AWS_ACCESS_KEY_ID',
                    secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'
                ]]) {
                    sh 'terraform -chdir=terraform plan -out=tfplan -input=false'
                }
            }
        }

        stage('Apply') {
            when {
                expression { env.GIT_BRANCH?.endsWith('/dev') }
            }
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId   : "${AWS_CREDENTIALS_ID}",
                    accessKeyVariable: 'AWS_ACCESS_KEY_ID',
                    secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'
                ]]) {
                    sh 'terraform -chdir=terraform apply -auto-approve -input=false tfplan'
                }
            }
        }

        stage('Wait Before Destroy') {
            when {
                expression { env.GIT_BRANCH?.endsWith('/dev') }
            }
            steps {
                echo 'Waiting 2 minutes before destroying resources...'
                sleep time: 2, unit: 'MINUTES'
            }
        }

        stage('Destroy') {
            when {
                expression { env.GIT_BRANCH?.endsWith('/dev') }
            }
            steps {
                withCredentials([[$class: 'AmazonWebServicesCredentialsBinding',
                    credentialsId   : "${AWS_CREDENTIALS_ID}",
                    accessKeyVariable: 'AWS_ACCESS_KEY_ID',
                    secretKeyVariable: 'AWS_SECRET_ACCESS_KEY'
                ]]) {
                    sh 'terraform -chdir=terraform destroy -auto-approve'
                }
            }
        }
    }

    post {
        always {
            sh 'rm -f terraform/tfplan'
        }
    }
}
