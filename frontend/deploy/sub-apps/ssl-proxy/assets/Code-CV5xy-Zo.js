import{_ as q,r as m,aj as F,f as I,a4 as j,i as V,a5 as w,j as D,d as O,k as N,o as _,c as y,l as z,n as g,m as J,p as K,x as U,J as b,q as A,D as G,ah as S}from"./index-BNxX6om6.js";function Q(n,e){const l=q(F,null);return m(()=>n.hljs||(l==null?void 0:l.mergedHljsRef.value))}function X(n){const{textColor2:e,fontSize:l,fontWeightStrong:a,textColor3:c}=n;return{textColor:e,fontSize:l,fontWeightStrong:a,"mono-3":"#a0a1a7","hue-1":"#0184bb","hue-2":"#4078f2","hue-3":"#a626a4","hue-4":"#50a14f","hue-5":"#e45649","hue-5-2":"#c91243","hue-6":"#986801","hue-6-2":"#c18401",lineNumberTextColor:c}}const Y={common:I,self:X};var Z=j([V("code",`
 font-size: var(--n-font-size);
 font-family: var(--n-font-family);
 `,[w("show-line-numbers",`
 display: flex;
 `),D("line-numbers",`
 user-select: none;
 padding-right: 12px;
 text-align: right;
 transition: color .3s var(--n-bezier);
 color: var(--n-line-number-text-color);
 `),w("word-wrap",[j("pre",`
 white-space: pre-wrap;
 word-break: break-all;
 `)]),j("pre",`
 margin: 0;
 line-height: inherit;
 font-size: inherit;
 font-family: inherit;
 `),j("[class^=hljs]",`
 color: var(--n-text-color);
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),({props:n})=>{const e=`${n.bPrefix}code`;return[`${e} .hljs-comment,
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
 }`]}]);const ee={...N.props,language:String,code:{type:String,default:""},trim:{type:Boolean,default:!0},hljs:Object,uri:Boolean,inline:Boolean,wordWrap:Boolean,showLineNumbers:Boolean,internalFontSize:Number,internalNoHighlight:Boolean};var te=O({name:"Code",props:ee,setup(n,{slots:e}){const{internalNoHighlight:l}=n,{mergedClsPrefixRef:a,inlineThemeDisabled:c}=K(),u=G(null),v=l?{value:void 0}:Q(n),R=(t,r,o)=>{const{value:s}=v;return!s||!(t&&s.getLanguage(t))?null:s.highlight(o?r.trim():r,{language:t}).value},p=m(()=>n.inline||n.wordWrap?!1:n.showLineNumbers),f=()=>{if(e.default)return;const{value:t}=u;if(!t)return;const{language:r}=n,o=n.uri?window.decodeURIComponent(n.code):n.code;if(r){const i=R(r,o,n.trim);if(i!==null){if(n.inline)t.innerHTML=i;else{const $=t.querySelector(".__code__");$&&t.removeChild($);const d=document.createElement("pre");d.className="__code__",d.innerHTML=i,t.appendChild(d)}return}}if(n.inline){t.textContent=o;return}const s=t.querySelector(".__code__");if(s)s.textContent=o;else{const i=document.createElement("pre");i.className="__code__",i.textContent=o,t.innerHTML="",t.appendChild(i)}};U(f),b(S(n,"language"),f),b(S(n,"code"),f),l||b(v,f);const L=N("Code","-code",Z,Y,n,a),x=m(()=>{const{common:{cubicBezierEaseInOut:t,fontFamilyMono:r},self:{textColor:o,fontSize:s,fontWeightStrong:i,lineNumberTextColor:$,"mono-3":d,"hue-1":k,"hue-2":B,"hue-3":E,"hue-4":T,"hue-5":P,"hue-5-2":H,"hue-6":M,"hue-6-2":W}}=L.value,{internalFontSize:C}=n;return{"--n-font-size":C?`${C}px`:s,"--n-font-family":r,"--n-font-weight-strong":i,"--n-bezier":t,"--n-text-color":o,"--n-mono-3":d,"--n-hue-1":k,"--n-hue-2":B,"--n-hue-3":E,"--n-hue-4":T,"--n-hue-5":P,"--n-hue-5-2":H,"--n-hue-6":M,"--n-hue-6-2":W,"--n-line-number-text-color":$}}),h=c?A("code",m(()=>`${n.internalFontSize||"a"}`),x,n):void 0;return{mergedClsPrefix:a,codeRef:u,mergedShowLineNumbers:p,lineNumbers:m(()=>{let t=1;const r=[];let o=!1;for(const s of n.code)s===`
`?(o=!0,r.push(t++)):o=!1;return o||r.push(t++),r.join(`
`)}),cssVars:c?void 0:x,themeClass:h==null?void 0:h.themeClass,onRender:h==null?void 0:h.onRender}},render(){const{mergedClsPrefix:n,wordWrap:e,mergedShowLineNumbers:l,onRender:a}=this;return a==null||a(),_(),y("code",{class:z([`${n}-code`,this.themeClass,e&&`${n}-code--word-wrap`,l&&`${n}-code--show-line-numbers`]),style:J(this.cssVars),ref:"codeRef"},[l?(_(),y("pre",{key:0,class:z(`${n}-code__line-numbers`)},[g(()=>this.lineNumbers)],2)):g(()=>null),g(()=>{var c,u;return(u=(c=this.$slots).default)==null?void 0:u.call(c)})],6)}});export{te as C};
