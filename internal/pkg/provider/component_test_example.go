/**
 * Copyright 2020 Napptive
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package provider

// Component example.
// Found in https://github.com/oam-dev/samples/blob/master/2.ServiceTracker_App/Components/tracker-data-component.yaml
var component=[]byte( `
apiVersion: core.oam.dev/v1alpha2
kind: Component
metadata:
  name: data-api
spec:
  workload:
    apiVersion: core.oam.dev/v1alpha2
    kind: ContainerizedWorkload
    metadata:
      name: data-api
    spec:
      osType: linux
      arch: amd64
      containers:
        - name: data-api
          image: artursouza/rudr-data-api:0.50
          env:
            - name: DATABASE_USER
              fromSecret:
                name: postgresqlconn
                key: username
            - name: DATABASE_PASSWORD
              fromSecret:
                name: postgresqlconn
                key: password
            - name: DATABASE_HOSTNAME
              fromSecret:
                name: postgresqlconn
                key: endpoint
            - name: DATABASE_NAME
              value: postgres
            - name: DATABASE_PORT
              value: 5432 
            - name: DATABASE_DRIVER
              value: postgres    
            - name: DATABASE_OPTIONS
              value: ""
          ports:
            - name: http
              containerPort: 3009
              protocol: TCP
          readinessProbe:
            exec:
              command:
                - wget
                - -q
                - 'http://127.0.0.1:3009/status'
                - -O
                - /dev/null
                - -S
            failureThreshold: 6
            initialDelaySeconds: 5
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5
          livenessProbe:
            exec:
              command:
                - wget
                - -q
                - 'http://127.0.0.1:3009/status'
                - -O
                - /dev/null
                - -S
            failureThreshold: 6
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5 
  parameters:
    - name: dbsecret
      description: secret with database connection information
      required: false
      fieldPaths:
      - spec.containers[0].env[0].fromSecret.name
      - spec.containers[0].env[1].fromSecret.name
      - spec.containers[0].env[2].fromSecret.name
    - name: dbname
      description: database name
      required: false
      fieldPaths:
      - spec.containers[0].env[3].value
    - name: dbport
      description: database port number
      required: false
      fieldPaths:
      - spec.containers[0].env[4].value
    - name: dbdriver
      description: database driver - one of 'mysql' | 'mariadb' | 'postgres' | 'mssql'
      required: false
      fieldPaths:
      - spec.containers[0].env[5].value
    - name: dboptions
      description: config as JSON
      required: false
      fieldPaths:
      - spec.containers[0].env[6].value
`)

