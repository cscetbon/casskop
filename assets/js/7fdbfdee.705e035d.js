"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[8202],{3503:(e,s,r)=>{r.r(s),r.d(s,{assets:()=>a,contentTitle:()=>o,default:()=>u,frontMatter:()=>t,metadata:()=>c,toc:()=>l});var n=r(5893),i=r(1151);const t={title:"Cassandra Cluster",sidebar_label:"Cassandra Cluster"},o=void 0,c={id:"configuration_deployment/cassandra_cluster",title:"Cassandra Cluster",description:"The full schema of the CassandraCluster resource is described in the Cassandra Cluster CRD Definition.",source:"@site/docs/3_configuration_deployment/2_cassandra_cluster.md",sourceDirName:"3_configuration_deployment",slug:"/configuration_deployment/cassandra_cluster",permalink:"/casskop/docs/configuration_deployment/cassandra_cluster",draft:!1,unlisted:!1,editUrl:"https://github.com/cscetbon/casskop/edit/master/website/docs/3_configuration_deployment/2_cassandra_cluster.md",tags:[],version:"current",sidebarPosition:2,frontMatter:{title:"Cassandra Cluster",sidebar_label:"Cassandra Cluster"},sidebar:"docs",previous:{title:"Customizable install with Helm",permalink:"/casskop/docs/configuration_deployment/customizable_install_with_helm"},next:{title:"Storage",permalink:"/casskop/docs/configuration_deployment/storage"}},a={},l=[{value:"Resource limits and requests",id:"resource-limits-and-requests",level:2},{value:"Resource requests",id:"resource-requests",level:3},{value:"Resource limits",id:"resource-limits",level:3},{value:"Supported CPU formats",id:"supported-cpu-formats",level:3},{value:"Supported memory formats",id:"supported-memory-formats",level:3},{value:"Configuring resource requests and limits",id:"configuring-resource-requests-and-limits",level:2}];function d(e){const s={a:"a",admonition:"admonition",code:"code",h2:"h2",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,i.a)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsxs)(s.p,{children:["The full schema of the ",(0,n.jsx)(s.code,{children:"CassandraCluster"})," resource is described in the ",(0,n.jsx)(s.a,{href:"#cassandra-cluster-crd-definition-version-020",children:"Cassandra Cluster CRD Definition"}),"."]}),"\n",(0,n.jsxs)(s.p,{children:["All labels that are applied to the desired ",(0,n.jsx)(s.code,{children:"CassandraCluster"})," resource will also be applied to the Kubernetes resources\nmaking up the Cassandra cluster. This provides a convenient mechanism for those resources to be labelled in whatever way\nthe user requires."]}),"\n",(0,n.jsx)(s.p,{children:"For every deployed container, CassKop allows you to specify the resources which should be reserved for it\nand the maximum resources that can be consumed by it. We support two types of resources:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsx)(s.li,{children:"Memory"}),"\n",(0,n.jsx)(s.li,{children:"CPU"}),"\n"]}),"\n",(0,n.jsx)(s.p,{children:"CassKop is using the Kubernetes syntax for specifying CPU and memory resources."}),"\n",(0,n.jsx)(s.h2,{id:"resource-limits-and-requests",children:"Resource limits and requests"}),"\n",(0,n.jsxs)(s.p,{children:["Resource limits and requests can be configured using the ",(0,n.jsx)(s.code,{children:"resources"})," property in ",(0,n.jsx)(s.code,{children:"CassandraCluster.spec.resources"}),"."]}),"\n",(0,n.jsx)(s.h3,{id:"resource-requests",children:"Resource requests"}),"\n",(0,n.jsx)(s.p,{children:"Requests specify the resources"}),"\n",(0,n.jsx)(s.admonition,{type:"important",children:(0,n.jsx)(s.p,{children:'If the resource request is for more than the available free resources on the scheduled kubernetes node,\nthe pod will remain stuck in "pending" state until the required resources become available.'})}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-yaml",children:"# ...\nresources:\n  requests:\n    cpu: 12\n    memory: 64Gi\n# ...\n"})}),"\n",(0,n.jsx)(s.h3,{id:"resource-limits",children:"Resource limits"}),"\n",(0,n.jsx)(s.p,{children:"Limits specify the maximum resources that can be consumed by a given container. The limit is not reserved and might not\nbe always available. The container can use the resources up to the limit only when they are available. The resource\nlimits should be always higher than the resource requests. If you only set limits, k8s uses the same value to set\nrequests."}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-yaml",children:"# ...\nresources:\n  limits:\n    cpu: 12\n    memory: 64Gi\n# ...\n"})}),"\n",(0,n.jsx)(s.h3,{id:"supported-cpu-formats",children:"Supported CPU formats"}),"\n",(0,n.jsx)(s.p,{children:"CPU requests and limits are supported in the following formats:"}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsxs)(s.li,{children:["Number of CPU cores as integer (",(0,n.jsx)(s.code,{children:"5"})," CPU core) or decimal (",(0,n.jsx)(s.code,{children:"2.5"}),"CPU core)."]}),"\n",(0,n.jsxs)(s.li,{children:["Number of millicpus / millicores (",(0,n.jsx)(s.code,{children:"100m"}),") where 1000 millicores is the same as ",(0,n.jsx)(s.code,{children:"1"})," CPU core."]}),"\n"]}),"\n",(0,n.jsxs)(s.p,{children:["For more details about CPU specification, refer to\n",(0,n.jsx)(s.a,{href:"https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu",children:"kubernetes documentation"})]}),"\n",(0,n.jsx)(s.h3,{id:"supported-memory-formats",children:"Supported memory formats"}),"\n",(0,n.jsx)(s.p,{children:"Memory requests and limits are specified in megabytes, gigabytes, mebibytes, gibibytes."}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsxs)(s.li,{children:["to specify memory in megabytes, use the ",(0,n.jsx)(s.code,{children:"M"})," suffix. For example ",(0,n.jsx)(s.code,{children:"1000M"}),"."]}),"\n",(0,n.jsxs)(s.li,{children:["to specify memory in gigabytes, use the ",(0,n.jsx)(s.code,{children:"G"})," suffix. For example ",(0,n.jsx)(s.code,{children:"1G"}),"."]}),"\n",(0,n.jsxs)(s.li,{children:["to specify memory in mebibytes, use the ",(0,n.jsx)(s.code,{children:"Mi"})," suffix. For example ",(0,n.jsx)(s.code,{children:"1000Mi"}),"."]}),"\n",(0,n.jsxs)(s.li,{children:["to specify memory in gibibytes, use the ",(0,n.jsx)(s.code,{children:"Gi"})," suffix. For example ",(0,n.jsx)(s.code,{children:"1Gi"}),"."]}),"\n"]}),"\n",(0,n.jsxs)(s.p,{children:["For more details about CPU specification, refer to\n",(0,n.jsx)(s.a,{href:"https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-memory",children:"kubernetes documentation"})]}),"\n",(0,n.jsx)(s.h2,{id:"configuring-resource-requests-and-limits",children:"Configuring resource requests and limits"}),"\n",(0,n.jsx)(s.p,{children:"the resources requests and limits for CPU and memory will be applied to all Cassandra Pods deployed in the Cluster."}),"\n",(0,n.jsxs)(s.p,{children:["It is configured directly in the ",(0,n.jsx)(s.code,{children:"CassandraCluster.spec.resources"}),":"]}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-yaml",children:"  resources:\n    requests:\n      cpu: '1'\n      memory: 1Gi\n    limits:\n      cpu: '2'\n      memory: 2Gi\n"})}),"\n",(0,n.jsx)(s.p,{children:"Depending on the values specified, Kubernetes will define 3 levels for QoS : (BestEffort < Burstable < Guaranteed)."}),"\n",(0,n.jsxs)(s.ul,{children:["\n",(0,n.jsx)(s.li,{children:"BestEffort: if no resources are specified"}),"\n",(0,n.jsx)(s.li,{children:"Burstable: if limits > requests. if a system needs more resources, thoses pods can be terminated if they use more than\nrequested and if there is no more BestEffort Pods to terminated"}),"\n",(0,n.jsx)(s.li,{children:"Guaranteed: request=limits. It is the recommended configuration for cassandra pods."}),"\n"]}),"\n",(0,n.jsxs)(s.p,{children:["When updating the crd resources, this will trigger an ",(0,n.jsx)(s.a,{href:"/casskop/docs/operations/cluster_operations#updateresources",children:"UpdateResources"})," action."]})]})}function u(e={}){const{wrapper:s}={...(0,i.a)(),...e.components};return s?(0,n.jsx)(s,{...e,children:(0,n.jsx)(d,{...e})}):d(e)}},1151:(e,s,r)=>{r.d(s,{Z:()=>c,a:()=>o});var n=r(7294);const i={},t=n.createContext(i);function o(e){const s=n.useContext(t);return n.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function c(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(i):e.components||i:o(e.components),n.createElement(t.Provider,{value:s},e.children)}}}]);