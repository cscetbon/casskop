"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[4628],{2240:(e,n,o)=>{o.r(n),o.d(n,{assets:()=>l,contentTitle:()=>r,default:()=>u,frontMatter:()=>i,metadata:()=>a,toc:()=>c});var s=o(5893),t=o(1151);const i={title:"GKE Issues",sidebar_label:"GKE Issues"},r=void 0,a={id:"troubleshooting/gke_issues",title:"GKE Issues",description:"RBAC on Google Container Engine (GKE)",source:"@site/docs/7_troubleshooting/2_gke_issues.md",sourceDirName:"7_troubleshooting",slug:"/troubleshooting/gke_issues",permalink:"/casskop/docs/troubleshooting/gke_issues",draft:!1,unlisted:!1,editUrl:"https://github.com/cscetbon/casskop/edit/master/website/docs/7_troubleshooting/2_gke_issues.md",tags:[],version:"current",sidebarPosition:2,frontMatter:{title:"GKE Issues",sidebar_label:"GKE Issues"},sidebar:"docs",previous:{title:"Operations Issues",permalink:"/casskop/docs/troubleshooting/operations_issues"},next:{title:"Developer guide",permalink:"/casskop/docs/contributing/developer_guide"}},l={},c=[{value:"RBAC on Google Container Engine (GKE)",id:"rbac-on-google-container-engine-gke",level:3},{value:"Pod and volumes can be scheduled in different zones using default provisioned",id:"pod-and-volumes-can-be-scheduled-in-different-zones-using-default-provisioned",level:3}];function d(e){const n={a:"a",admonition:"admonition",code:"code",h3:"h3",li:"li",p:"p",pre:"pre",ul:"ul",...(0,t.a)(),...e.components};return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(n.h3,{id:"rbac-on-google-container-engine-gke",children:"RBAC on Google Container Engine (GKE)"}),"\n",(0,s.jsxs)(n.p,{children:["When you try to create ",(0,s.jsx)(n.code,{children:"ClusterRole"})," (",(0,s.jsx)(n.code,{children:"casskop"}),", etc.) on GKE Kubernetes cluster, you will probably run into permission errors:"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{children:'<....>\nfailed to initialize cluster resources: roles.rbac.authorization.k8s.io\n"casskop" is forbidden: attempt to grant extra privileges:\n<....>\n'})}),"\n",(0,s.jsxs)(n.p,{children:["This is due to the way Container Engine checks permissions. From ",(0,s.jsx)(n.a,{href:"https://cloud.google.com/container-engine/docs/role-based-access-control",children:"Google Container Engine docs"}),":"]}),"\n",(0,s.jsx)(n.admonition,{type:"note",children:(0,s.jsx)(n.p,{children:"Because of the way Container Engine checks permissions when you create a Role or ClusterRole, you must first create a RoleBinding that grants you all of the permissions included in the role you want to create.\nAn example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.\nThis is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6."})}),"\n",(0,s.jsxs)(n.p,{children:["To overcome this, you must grant your current Google identity ",(0,s.jsx)(n.code,{children:"cluster-admin"})," Role:"]}),"\n",(0,s.jsx)(n.pre,{children:(0,s.jsx)(n.code,{className:"language-console",children:'# get current google identity\n$ gcloud info | grep Account\nAccount: [myname@example.org]\n\n# grant cluster-admin to your current identity\n$ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org\nClusterrolebinding "myname-cluster-admin-binding" created\n'})}),"\n",(0,s.jsx)(n.h3,{id:"pod-and-volumes-can-be-scheduled-in-different-zones-using-default-provisioned",children:"Pod and volumes can be scheduled in different zones using default provisioned"}),"\n",(0,s.jsxs)(n.p,{children:["The default provisioner in GKE does not have the ",(0,s.jsx)(n.code,{children:'volumeBindingMode: "WaitForFirstConsumer"'})," option that can result in\na bad\nscheduling behaviour.\nWe use one of the following files to create a storage class:"]}),"\n",(0,s.jsxs)(n.ul,{children:["\n",(0,s.jsx)(n.li,{children:"config/samples/gke-storage-standard-wait.yaml"}),"\n",(0,s.jsx)(n.li,{children:"config/samples/gke-storage-ssd-wait.yaml (if you have ssd disks)"}),"\n"]})]})}function u(e={}){const{wrapper:n}={...(0,t.a)(),...e.components};return n?(0,s.jsx)(n,{...e,children:(0,s.jsx)(d,{...e})}):d(e)}},1151:(e,n,o)=>{o.d(n,{Z:()=>a,a:()=>r});var s=o(7294);const t={},i=s.createContext(t);function r(e){const n=s.useContext(i);return s.useMemo((function(){return"function"==typeof e?e(n):{...n,...e}}),[n,e])}function a(e){let n;return n=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:r(e.components),s.createElement(i.Provider,{value:n},e.children)}}}]);