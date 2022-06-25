pipeline {
    agent any
    
    stages {
        stage('Build') {
            steps {
                checkout poll: false, scm: [$class: 'GitSCM', branches: [[name: '*/master']], extensions: [], userRemoteConfigs: [[credentialsId: 'scm-manager-mandrakey', url: 'https://scm.bleuelmedia.com/scm/repo/shoptrac/backend']]]
                sh "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build"
                archiveArtifacts artifacts: 'shoptrac'
            }
        }
        stage('Deploy to production') {
            sshPublisher(publishers: [sshPublisherDesc(configName: 'agate.top-web.info', transfers: [sshTransfer(cleanRemote: false, excludes: '', execCommand: 'cp /opt/shoptrac/shoptrac /opt/shoptrac/shoptrac.bak && chmod 0775 /tmp/shoptrac && mv /tmp/shoptrac /opt/shoptrac/shoptrac && sudo systemctl restart shoptrac', execTimeout: 120000, flatten: false, makeEmptyDirs: false, noDefaultExcludes: false, patternSeparator: '[, ]+', remoteDirectory: '', remoteDirectorySDF: false, removePrefix: '', sourceFiles: 'shoptrac')], usePromotionTimestamp: false, useWorkspaceInPromotion: false, verbose: false)])
        }
    }
}