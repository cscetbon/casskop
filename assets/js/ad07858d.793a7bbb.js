"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[3577],{8362:(n,e,a)=>{a.r(e),a.d(e,{assets:()=>c,contentTitle:()=>r,default:()=>u,frontMatter:()=>o,metadata:()=>t,toc:()=>l});var s=a(5893),i=a(1151);const o={title:"Sidecars",sidebar_label:"Sidecars"},r=void 0,t={id:"configuration_deployment/sidecars",title:"Sidecars",description:"For extra needs not covered by the defaults container managed through the CassandraCluster CRD, we are allowing you to define your own sidecars which will be deployed into the cassandra node pods.",source:"@site/docs/3_configuration_deployment/5_sidecars.md",sourceDirName:"3_configuration_deployment",slug:"/configuration_deployment/sidecars",permalink:"/casskop/docs/configuration_deployment/sidecars",draft:!1,unlisted:!1,editUrl:"https://github.com/cscetbon/casskop/edit/master/website/docs/3_configuration_deployment/5_sidecars.md",tags:[],version:"current",sidebarPosition:5,frontMatter:{title:"Sidecars",sidebar_label:"Sidecars"},sidebar:"docs",previous:{title:"Cassandra Configuration",permalink:"/casskop/docs/configuration_deployment/cassandra_configuration"},next:{title:"Advanced Configuration",permalink:"/casskop/docs/configuration_deployment/advanced_configuration"}},c={},l=[];function d(n){const e={admonition:"admonition",code:"code",em:"em",li:"li",p:"p",pre:"pre",ul:"ul",...(0,i.a)(),...n.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsxs)(e.p,{children:["For extra needs not covered by the defaults container managed through the CassandraCluster CRD, we are allowing you to define your own sidecars which will be deployed into the cassandra node pods.\nTo do this, you will configure the ",(0,s.jsx)(e.code,{children:"SidecarConfigs"})," property in ",(0,s.jsx)(e.code,{children:"CassandraCluster.Spec"}),"."]}),"\n",(0,s.jsx)(e.p,{children:"CassandraCluster fragment for dynamic sidecars definition :"}),"\n",(0,s.jsx)(e.pre,{children:(0,s.jsx)(e.code,{className:"language-yaml",children:'# ...\n  sidecarConfigs:\n    - args: ["tail", "-F", "/var/log/cassandra/system.log"]\n      image: ez123/alpine-tini\n      imagePullPolicy: Always\n      name: cassandra-log\n      resources:\n        limits:\n          cpu: 50m\n          memory: 50Mi\n        requests:\n          cpu: 10m\n          memory: 10Mi\n      volumeMounts:\n        - mountPath: /var/log/cassandra\n          name: cassandra-logs\n    - args: ["tail", "-F", "/var/log/cassandra/gc.log.0.current"]\n      image: ez123/alpine-tini\n      imagePullPolicy: Always\n      name: gc-log\n      resources:\n        limits:\n          cpu: 50m\n          memory: 50Mi\n        requests:\n          cpu: 10m\n          memory: 10Mi\n      volumeMounts:\n        - mountPath: /var/log/cassandra\n          name: gc-logs\n# ...\n'})}),"\n",(0,s.jsxs)(e.ul,{children:["\n",(0,s.jsxs)(e.li,{children:[(0,s.jsx)(e.code,{children:"sidecarConfigs"})," ",(0,s.jsx)(e.em,{children:"(required)"})," : Defines the list of container config object, which will be added into each pod of cassandra node, it requires a list of kubernetes Container spec."]}),"\n"]}),"\n",(0,s.jsxs)(e.p,{children:["With the above configuration, the following configuration will be added to the ",(0,s.jsx)(e.code,{children:"rack statefulset"})," definition :"]}),"\n",(0,s.jsx)(e.pre,{children:(0,s.jsx)(e.code,{className:"language-yaml",children:'# ...\n#   ...\n  containers:\n    - args: ["tail", "-F", "/var/log/cassandra/system.log"]\n      image: ez123/alpine-tini\n      imagePullPolicy: Always\n      name: cassandra-log\n      resources:\n        limits:\n          cpu: 50m\n          memory: 50Mi\n        requests:\n          cpu: 10m\n          memory: 10Mi\n      volumeMounts:\n        - mountPath: /var/log/cassandra\n          name: cassandra-logs\n    - args: ["tail", "-F", "/var/log/cassandra/gc.log.0.current"]\n      image: ez123/alpine-tini\n      imagePullPolicy: Always\n      name: gc-log\n      resources:\n        limits:\n          cpu: 50m\n          memory: 50Mi\n        requests:\n          cpu: 10m\n          memory: 10Mi\n      volumeMounts:\n        - mountPath: /var/log/cassandra\n          name: gc-logs\n#   ...\n# ...\n'})}),"\n",(0,s.jsxs)(e.admonition,{type:"info",children:[(0,s.jsx)(e.p,{children:"Note that all sidecars added with this configuration will have some of the environment variables from cassandra container merged with those defined into the sidecar container\nfor example :"}),(0,s.jsxs)(e.ul,{children:["\n",(0,s.jsx)(e.li,{children:"CASSANDRA_CLUSTER_NAME"}),"\n",(0,s.jsx)(e.li,{children:"CASSANDRA_SEEDS"}),"\n",(0,s.jsx)(e.li,{children:"CASSANDRA_DC"}),"\n",(0,s.jsx)(e.li,{children:"CASSANDRA_RACK"}),"\n"]})]})]})}function u(n={}){const{wrapper:e}={...(0,i.a)(),...n.components};return e?(0,s.jsx)(e,{...n,children:(0,s.jsx)(d,{...n})}):d(n)}},1151:(n,e,a)=>{a.d(e,{Z:()=>t,a:()=>r});var s=a(7294);const i={},o=s.createContext(i);function r(n){const e=s.useContext(o);return s.useMemo((function(){return"function"==typeof n?n(e):{...e,...n}}),[e,n])}function t(n){let e;return e=n.disableParentContext?"function"==typeof n.components?n.components(i):n.components||i:r(n.components),s.createElement(o.Provider,{value:e},n.children)}}}]);