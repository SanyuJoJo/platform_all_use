import{a as z,S as J,E as K}from"./Spin-QgI_6sok.js";import{J as Q,r as m,ag as U,f as Y,Q as b,i as Z,R as N,j as A,d as T,k as E,o as g,c as y,l as B,n as _,m as G,p as X,Z as ee,Y as x,q as te,D as ne,a8 as k,a as w,u as d,w as v,h as oe,ab as le,B as se,e as re,s as R}from"./index-CVUczFfF.js";import{u as ae}from"./use-message-D3mBhG8n.js";import{_ as ie}from"./_plugin-vue_export-helper-DlAUqK2U.js";function ce(t,e){const o=Q(U,null);return m(()=>t.hljs||(o==null?void 0:o.mergedHljsRef.value))}function ue(t){const{textColor2:e,fontSize:o,fontWeightStrong:r,textColor3:c}=t;return{textColor:e,fontSize:o,fontWeightStrong:r,"mono-3":"#a0a1a7","hue-1":"#0184bb","hue-2":"#4078f2","hue-3":"#a626a4","hue-4":"#50a14f","hue-5":"#e45649","hue-5-2":"#c91243","hue-6":"#986801","hue-6-2":"#c18401",lineNumberTextColor:c}}const he={common:Y,self:ue};var de=b([Z("code",`
 font-size: var(--n-font-size);
 font-family: var(--n-font-family);
 `,[N("show-line-numbers",`
 display: flex;
 `),A("line-numbers",`
 user-select: none;
 padding-right: 12px;
 text-align: right;
 transition: color .3s var(--n-bezier);
 color: var(--n-line-number-text-color);
 `),N("word-wrap",[b("pre",`
 white-space: pre-wrap;
 word-break: break-all;
 `)]),b("pre",`
 margin: 0;
 line-height: inherit;
 font-size: inherit;
 font-family: inherit;
 `),b("[class^=hljs]",`
 color: var(--n-text-color);
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),({props:t})=>{const e=`${t.bPrefix}code`;return[`${e} .hljs-comment,
 ${e} .hljs-quote {
 color: var(--n-mono-3);
 font-style: italic;
 }`,`${e} .hljs-doctag,
 ${e} .hljs-keyword,
 ${e} .hljs-formula {
 color: var(--n-hue-3);
 }`,`${e} .hljs-section,
 ${e} .hljs-name,
 ${e} .hljs-selector-tag,
 ${e} .hljs-deletion,
 ${e} .hljs-subst {
 color: var(--n-hue-5);
 }`,`${e} .hljs-literal {
 color: var(--n-hue-1);
 }`,`${e} .hljs-string,
 ${e} .hljs-regexp,
 ${e} .hljs-addition,
 ${e} .hljs-attribute,
 ${e} .hljs-meta-string {
 color: var(--n-hue-4);
 }`,`${e} .hljs-built_in,
 ${e} .hljs-class .hljs-title {
 color: var(--n-hue-6-2);
 }`,`${e} .hljs-attr,
 ${e} .hljs-variable,
 ${e} .hljs-template-variable,
 ${e} .hljs-type,
 ${e} .hljs-selector-class,
 ${e} .hljs-selector-attr,
 ${e} .hljs-selector-pseudo,
 ${e} .hljs-number {
 color: var(--n-hue-6);
 }`,`${e} .hljs-symbol,
 ${e} .hljs-bullet,
 ${e} .hljs-link,
 ${e} .hljs-meta,
 ${e} .hljs-selector-id,
 ${e} .hljs-title {
 color: var(--n-hue-2);
 }`,`${e} .hljs-emphasis {
 font-style: italic;
 }`,`${e} .hljs-strong {
 font-weight: var(--n-font-weight-strong);
 }`,`${e} .hljs-link {
 text-decoration: underline;
 }`]}]);const fe={...E.props,language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean};var me=T({name:"Code",props:fe,setup(t,{slots:e}){const{internalNoHighlight:o}=t,{mergedClsPrefixRef:r,inlineThemeDisabled:c}=X(),u=ne(null),C=o?{value:void 0}:ce(t),L=(n,s,l)=>{const{value:a}=C;return!a||!(n&&a.getLanguage(n))?null:a.highlight(l?s.trim():s,{language:n}).value},M=m(()=>t.inline||t.wordWrap?!1:t.showLineNumbers),$=()=>{if(e.default)return;const{value:n}=u;if(!n)return;const{language:s}=t,l=t.uri?window.decodeURIComponent(t.code):t.code;if(s){const i=L(s,l,t.trim);if(i!==null){if(t.inline)n.innerHTML=i;else{const j=n.querySelector(".__code__");j&&n.removeChild(j);const f=document.createElement("pre");f.className="__code__",f.innerHTML=i,n.appendChild(f)}return}}if(t.inline){n.textContent=l;return}const a=n.querySelector(".__code__");if(a)a.textContent=l;else{const i=document.createElement("pre");i.className="__code__",i.textContent=l,n.innerHTML="",n.appendChild(i)}};ee($),x(k(t,"language"),$),x(k(t,"code"),$),o||x(C,$);const P=E("Code","-code",de,he,t,r),p=m(()=>{const{common:{cubicBezierEaseInOut:n,fontFamilyMono:s},self:{textColor:l,fontSize:a,fontWeightStrong:i,lineNumberTextColor:j,"mono-3":f,"hue-1":V,"hue-2":H,"hue-3":W,"hue-4":I,"hue-5":q,"hue-5-2":F,"hue-6":D,"hue-6-2":O}}=P.value,{internalFontSize:S}=t;return{"--n-font-size":S?`${S}px`:a,"--n-font-family":s,"--n-font-weight-strong":i,"--n-bezier":n,"--n-text-color":l,"--n-mono-3":f,"--n-hue-1":V,"--n-hue-2":H,"--n-hue-3":W,"--n-hue-4":I,"--n-hue-5":q,"--n-hue-5-2":F,"--n-hue-6":D,"--n-hue-6-2":O,"--n-line-number-text-color":j}}),h=c?te("code",m(()=>`${t.internalFontSize||"a"}`),p,t):void 0;return{mergedClsPrefix:r,codeRef:u,mergedShowLineNumbers:M,lineNumbers:m(()=>{let n=1;const s=[];let l=!1;for(const a of t.code)a===`
`?(l=!0,s.push(n++)):l=!1;return l||s.push(n++),s.join(`
`)}),cssVars:c?void 0:p,themeClass:h==null?void 0:h.themeClass,onRender:h==null?void 0:h.onRender}},render(){const{mergedClsPrefix:t,wordWrap:e,mergedShowLineNumbers:o,onRender:r}=this;return r==null||r(),g(),y("code",{class:B([`${t}-code`,this.themeClass,e&&`${t}-code--word-wrap`,o&&`${t}-code--show-line-numbers`]),style:G(this.cssVars),ref:"codeRef"},[o?(g(),y("pre",{key:0,class:B(`${t}-code__line-numbers`)},[_(()=>this.lineNumbers)],2)):_(()=>null),_(()=>{var c,u;return(u=(c=this.$slots).default)==null?void 0:u.call(c)})],6)}});const ge={class:"config-tab__title"},$e=T({__name:"ConfigTab",props:{title:{},content:{},loading:{type:Boolean}},setup(t){const e=t,o=ae();async function r(){if(!e.content){o.warning("没有可复制的内容");return}try{await navigator.clipboard.writeText(e.content),o.success("已复制到剪贴板")}catch{o.error("复制失败，请手动选择复制")}}return(c,u)=>(g(),y("div",null,[w(d(z),{justify:"space-between",align:"center",style:{"margin-bottom":"8px"}},{default:v(()=>[oe("span",ge,le(t.title),1),w(d(z),null,{default:v(()=>[w(d(se),{size:"small",onClick:r},{default:v(()=>[...u[0]||(u[0]=[re("复制",-1)])]),_:1})]),_:1})]),_:1}),w(d(J),{show:t.loading},{default:v(()=>[t.content?(g(),R(d(me),{key:0,code:t.content,language:"nginx","show-line-numbers":"","word-wrap":""},null,8,["code"])):(g(),R(d(K),{key:1,description:"暂无内容"}))]),_:1},8,["show"])]))}}),_e=ie($e,[["__scopeId","data-v-ffd331ee"]]);export{_e as default};
