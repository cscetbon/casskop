"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[8705],{2847:(e,s,n)=>{n.r(s),n.d(s,{assets:()=>c,contentTitle:()=>o,default:()=>p,frontMatter:()=>r,metadata:()=>i,toc:()=>d});var a=n(5893),t=n(1151);const r={title:"Introduction",sidebar_label:"Introduction"},o=void 0,i={id:"concepts/introduction",title:"Introduction",description:"The Orange Cassandra operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of Apache Cassandra clusters deployed to K8s.",source:"@site/docs/1_concepts/1_introduction.md",sourceDirName:"1_concepts",slug:"/concepts/introduction",permalink:"/casskop/docs/concepts/introduction",draft:!1,unlisted:!1,editUrl:"https://github.com/cscetbon/casskop/edit/master/website/docs/1_concepts/1_introduction.md",tags:[],version:"current",sidebarPosition:1,frontMatter:{title:"Introduction",sidebar_label:"Introduction"},sidebar:"docs",next:{title:"Design Principes",permalink:"/casskop/docs/concepts/design_principes"}},c={},d=[{value:"Overview",id:"overview",level:2},{value:"Presentation",id:"presentation",level:2},{value:"Motivation",id:"motivation",level:2}];function l(e){const s={a:"a",admonition:"admonition",h2:"h2",li:"li",p:"p",strong:"strong",ul:"ul",...(0,t.a)(),...e.components};return(0,a.jsxs)(a.Fragment,{children:[(0,a.jsxs)(s.p,{children:["The Orange Cassandra operator is a Kubernetes operator to automate provisioning, management, autoscaling and operations of ",(0,a.jsx)(s.a,{href:"http://cassandra.apache.org/",children:"Apache Cassandra"})," clusters deployed to K8s."]}),"\n",(0,a.jsx)(s.h2,{id:"overview",children:"Overview"}),"\n",(0,a.jsx)(s.p,{children:"The CassKop Cassandra Kubernetes operator makes it easy to run Apache Cassandra on Kubernetes. Apache Cassandra is a popular,\nfree, open-source, distributed wide column store, NoSQL database management system.\nThe operator allows to easily create and manage racks and data centers aware Cassandra clusters."}),"\n",(0,a.jsxs)(s.p,{children:["Some of the high-level capabilities and objectives of Apache Cassandra include, and some of the main features of the ",(0,a.jsx)(s.strong,{children:"Casskop"})," are:"]}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsx)(s.li,{children:"Deployment of a C* cluster (rack or AZ aware)"}),"\n",(0,a.jsx)(s.li,{children:"Graceful rolling update"}),"\n",(0,a.jsxs)(s.li,{children:["Graceful C* cluster ",(0,a.jsx)(s.strong,{children:"scaling"})," (with cleanup and decommission prior to Kubernetes scale down)"]}),"\n",(0,a.jsx)(s.li,{children:"Manage operations on pods through CassKop plugin (cleanup, rebuild, upgradesstable, removenode..)"}),"\n",(0,a.jsxs)(s.li,{children:["Performing live Cassandra repairs through the use of ",(0,a.jsx)(s.a,{href:"http://cassandra-reaper.io/",children:"Cassandra reaper"})]}),"\n",(0,a.jsxs)(s.li,{children:["Multi-site management through ",(0,a.jsx)(s.a,{href:"https://github.com/cscetbon/casskop/tree/master/multi-casskop",children:"Multi-Casskop operator"})]}),"\n",(0,a.jsx)(s.li,{children:"Live Backup/Restore of Cassandra's datas"}),"\n"]}),"\n",(0,a.jsxs)(s.p,{children:["The Cassandra operator is based on the CoreOS\n",(0,a.jsx)(s.a,{href:"https://github.com/operator-framework/operator-sdk",children:"operator-sdk"})," tools and APIs."]}),"\n",(0,a.jsxs)(s.p,{children:["CassKop creates/configures/manages Cassandra clusters atop Kubernetes and is by default ",(0,a.jsx)(s.strong,{children:"space-scoped"})," which means\nthat :"]}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsx)(s.li,{children:"CassKop is able to manage X Cassandra clusters in one Kubernetes namespace."}),"\n",(0,a.jsx)(s.li,{children:"You need X instances of CassKop to manage Y Cassandra clusters in X different namespaces (1 instance of CassKop\nper namespace)."}),"\n"]}),"\n",(0,a.jsx)(s.admonition,{type:"info",children:(0,a.jsx)(s.p,{children:"This adds security between namespaces with a better isolation, and less work for each operator."})}),"\n",(0,a.jsx)(s.h2,{id:"presentation",children:"Presentation"}),"\n",(0,a.jsxs)(s.p,{children:["We have some slides for a ",(0,a.jsx)(s.a,{href:"https://cscetbon.github.io/casskop/slides/index.html?slides=Slides-CassKop-demo.md#1",children:"CassKop demo"})]}),"\n",(0,a.jsxs)(s.p,{children:["You can also play with CassKop on ",(0,a.jsx)(s.a,{href:"https://www.katacoda.com/orange",children:"Katacoda"})]}),"\n",(0,a.jsx)(s.h2,{id:"motivation",children:"Motivation"}),"\n",(0,a.jsxs)(s.p,{children:["At ",(0,a.jsx)(s.a,{href:"https://opensource.orange.com/fr/accueil/",children:"Orange"})," we are building some ",(0,a.jsx)(s.a,{href:"https://github.com/cscetbon?utf8=%E2%9C%93&q=operator&type=&language=",children:"Kubernetes operator"}),", that operate NiFi, Galera and Cassandra clusters (among other types) for our business cases."]}),"\n",(0,a.jsx)(s.p,{children:"There are already some approaches to operating C* on Kubernetes, however, we did not find them appropriate for use in a highly dynamic environment, nor capable of meeting our needs."}),"\n",(0,a.jsxs)(s.ul,{children:["\n",(0,a.jsxs)(s.li,{children:[(0,a.jsx)(s.a,{href:"https://github.com/k8ssandra/cass-operator",children:"Datastax K8ssandra Cass-Operator"})," (see also ",(0,a.jsx)(s.a,{href:"https://k8ssandra.io",children:"K8ssandra project"}),")"]}),"\n",(0,a.jsx)(s.li,{children:(0,a.jsx)(s.a,{href:"https://github.com/instaclustr/cassandra-operator",children:"Instaclustr Operator"})}),"\n",(0,a.jsx)(s.li,{children:(0,a.jsx)(s.a,{href:"https://github.com/sky-uk/cassandra-operator",children:"Sky-Uk Operator"})}),"\n"]}),"\n",(0,a.jsx)(s.p,{children:"Finally, our motivation is to build an open source solution and a community which drives the innovation and features of this operator."})]})}function p(e={}){const{wrapper:s}={...(0,t.a)(),...e.components};return s?(0,a.jsx)(s,{...e,children:(0,a.jsx)(l,{...e})}):l(e)}},1151:(e,s,n)=>{n.d(s,{Z:()=>i,a:()=>o});var a=n(7294);const t={},r=a.createContext(t);function o(e){const s=a.useContext(r);return a.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function i(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:o(e.components),a.createElement(r.Provider,{value:s},e.children)}}}]);