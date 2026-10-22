import{f as b,i as C,a3 as h,j as n,J as t,d as $,k as u,o as g,c as m,l as v,n as f,U as _,h as p,m as w,p as k,q as z,r as y}from"./index-BG9A5pts.js";function B(i){const{textColor1:o,dividerColor:r,fontWeightStrong:l}=i;return{textColor:o,color:r,fontWeight:l}}const P={common:b,self:B};var R=C("divider",`
 position: relative;
 display: flex;
 width: 100%;
 box-sizing: border-box;
 font-size: 16px;
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
`,[h("vertical",`
 margin-top: 24px;
 margin-bottom: 24px;
 `,[h("no-title",`
 display: flex;
 align-items: center;
 `)]),n("title",`
 display: flex;
 align-items: center;
 margin-left: 12px;
 margin-right: 12px;
 white-space: nowrap;
 font-weight: var(--n-font-weight);
 `),t("title-position-left",[n("line",[t("left",{width:"28px"})])]),t("title-position-right",[n("line",[t("right",{width:"28px"})])]),t("dashed",[n("line",`
 background-color: #0000;
 height: 0px;
 width: 100%;
 border-style: dashed;
 border-width: 1px 0 0;
 `)]),t("vertical",`
 display: inline-block;
 height: 1em;
 margin: 0 8px;
 vertical-align: middle;
 width: 1px;
 `),n("line",`
 border: none;
 transition: background-color .3s var(--n-bezier), border-color .3s var(--n-bezier);
 height: 1px;
 width: 100%;
 margin: 0;
 `),h("dashed",[n("line",{backgroundColor:"var(--n-color)"})]),t("dashed",[n("line",{borderColor:"var(--n-color)"})]),t("vertical",{backgroundColor:"var(--n-color)"})]);const D={...u.props,titlePlacement:{type:String,default:"center"},dashed:Boolean,vertical:Boolean};var E=$({name:"Divider",props:D,setup(i){const{mergedClsPrefixRef:o,inlineThemeDisabled:r}=k(i),l=u("Divider","-divider",R,P,i,o),d=y(()=>{const{common:{cubicBezierEaseInOut:s},self:{color:a,textColor:c,fontWeight:x}}=l.value;return{"--n-bezier":s,"--n-color":a,"--n-text-color":c,"--n-font-weight":x}}),e=r?z("divider",void 0,d,i):void 0;return{mergedClsPrefix:o,cssVars:r?void 0:d,themeClass:e==null?void 0:e.themeClass,onRender:e==null?void 0:e.onRender}},render(){var s;const{$slots:i,titlePlacement:o,vertical:r,dashed:l,cssVars:d,mergedClsPrefix:e}=this;return(s=this.onRender)==null||s.call(this),g(),m("div",{role:"separator",class:v([`${e}-divider`,this.themeClass,{[`${e}-divider--vertical`]:r,[`${e}-divider--no-title`]:!i.default,[`${e}-divider--dashed`]:l,[`${e}-divider--title-position-${o}`]:i.default&&o}]),style:w(d)},[r?f(()=>null):(g(),m("div",{key:0,class:v(`${e}-divider__line ${e}-divider__line--left`)},null,2)),!r&&i.default?(g(),m(_,{key:2},[p("div",{class:v(`${e}-divider__title`)},[f(()=>{var a,c;return(c=(a=this.$slots).default)==null?void 0:c.call(a)})],2),p("div",{class:v(`${e}-divider__line ${e}-divider__line--right`)},null,2)],64)):f(()=>null)],6)}});export{E as D};
