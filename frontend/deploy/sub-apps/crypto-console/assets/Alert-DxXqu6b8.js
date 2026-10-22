import{f as D,G as u,ag as h,i as z,j as i,J as E,aX as G,H as J,d as K,k as H,o as s,s as p,aY as X,p as Y,M as Z,q as Q,r as A,A as U,c as _,n as v,aT as ee,l as y,a8 as oe,I as re,E as ne,W as te,v as le,S as se,h as ie,ak as ae,a9 as W,aZ as ce,t as c}from"./index-BG9A5pts.js";var de={iconMargin:"11px 8px 0 12px",iconMarginRtl:"11px 12px 0 8px",iconSize:"24px",closeIconSize:"16px",closeSize:"20px",closeMargin:"13px 14px 0 0",closeMarginRtl:"13px 0 0 14px",padding:"13px"};function fe(r){const{lineHeight:e,borderRadius:d,fontWeightStrong:x,baseColor:l,dividerColor:C,actionColor:k,textColor1:f,textColor2:n,closeColorHover:g,closeColorPressed:b,closeIconColor:m,closeIconColorHover:I,closeIconColorPressed:t,infoColor:o,successColor:S,warningColor:P,errorColor:T,fontSize:R}=r;return{...de,fontSize:R,lineHeight:e,titleFontWeight:x,borderRadius:d,border:`1px solid ${C}`,color:k,titleTextColor:f,iconColor:n,contentTextColor:n,closeBorderRadius:d,closeColorHover:g,closeColorPressed:b,closeIconColor:m,closeIconColorHover:I,closeIconColorPressed:t,borderInfo:`1px solid ${u(l,h(o,{alpha:.25}))}`,colorInfo:u(l,h(o,{alpha:.08})),titleTextColorInfo:f,iconColorInfo:o,contentTextColorInfo:n,closeColorHoverInfo:g,closeColorPressedInfo:b,closeIconColorInfo:m,closeIconColorHoverInfo:I,closeIconColorPressedInfo:t,borderSuccess:`1px solid ${u(l,h(S,{alpha:.25}))}`,colorSuccess:u(l,h(S,{alpha:.08})),titleTextColorSuccess:f,iconColorSuccess:S,contentTextColorSuccess:n,closeColorHoverSuccess:g,closeColorPressedSuccess:b,closeIconColorSuccess:m,closeIconColorHoverSuccess:I,closeIconColorPressedSuccess:t,borderWarning:`1px solid ${u(l,h(P,{alpha:.33}))}`,colorWarning:u(l,h(P,{alpha:.08})),titleTextColorWarning:f,iconColorWarning:P,contentTextColorWarning:n,closeColorHoverWarning:g,closeColorPressedWarning:b,closeIconColorWarning:m,closeIconColorHoverWarning:I,closeIconColorPressedWarning:t,borderError:`1px solid ${u(l,h(T,{alpha:.25}))}`,colorError:u(l,h(T,{alpha:.08})),titleTextColorError:f,iconColorError:T,contentTextColorError:n,closeColorHoverError:g,closeColorPressedError:b,closeIconColorError:m,closeIconColorHoverError:I,closeIconColorPressedError:t}}const ge={common:D,self:fe};var ue=z("alert",`
 line-height: var(--n-line-height);
 border-radius: var(--n-border-radius);
 position: relative;
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-color);
 text-align: start;
 word-break: break-word;
`,[i("border",`
 border-radius: inherit;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 transition: border-color .3s var(--n-bezier);
 border: var(--n-border);
 pointer-events: none;
 `),E("closable",[z("alert-body",[i("title",`
 padding-right: 24px;
 `)])]),i("icon",{color:"var(--n-icon-color)"}),z("alert-body",{padding:"var(--n-padding)"},[i("title",{color:"var(--n-title-text-color)"}),i("content",{color:"var(--n-content-text-color)"})]),G({originalTransition:"transform .3s var(--n-bezier)",enterToProps:{transform:"scale(1)"},leaveToProps:{transform:"scale(0.9)"}}),i("icon",`
 position: absolute;
 left: 0;
 top: 0;
 align-items: center;
 justify-content: center;
 display: flex;
 width: var(--n-icon-size);
 height: var(--n-icon-size);
 font-size: var(--n-icon-size);
 margin: var(--n-icon-margin);
 `),i("close",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 position: absolute;
 right: 0;
 top: 0;
 margin: var(--n-close-margin);
 `),E("show-icon",[z("alert-body",{paddingLeft:"calc(var(--n-icon-margin-left) + var(--n-icon-size) + var(--n-icon-margin-right))"})]),E("right-adjust",[z("alert-body",{paddingRight:"calc(var(--n-close-size) + var(--n-padding) + 2px)"})]),z("alert-body",`
 border-radius: var(--n-border-radius);
 transition: border-color .3s var(--n-bezier);
 `,[i("title",`
 transition: color .3s var(--n-bezier);
 font-size: 16px;
 line-height: 19px;
 font-weight: var(--n-title-font-weight);
 `,[J("& +",[i("content",{marginTop:"9px"})])]),i("content",{transition:"color .3s var(--n-bezier)",fontSize:"var(--n-font-size)"})]),i("icon",{transition:"color .3s var(--n-bezier)"})]);const he={...H.props,title:String,showIcon:{type:Boolean,default:!0},type:{type:String,default:"default"},bordered:{type:Boolean,default:!0},closable:Boolean,onClose:Function,onAfterLeave:Function,onAfterHide:Function};var Ce=K({name:"Alert",inheritAttrs:!1,props:he,slots:Object,setup(r){const{mergedClsPrefixRef:e,mergedBorderedRef:d,inlineThemeDisabled:x,mergedRtlRef:l}=Y(r),C=H("Alert","-alert",ue,ge,r,e),k=Z("Alert",l,e),f=A(()=>{const{common:{cubicBezierEaseInOut:t},self:o}=C.value,{fontSize:S,borderRadius:P,titleFontWeight:T,lineHeight:R,iconSize:w,iconMargin:$,iconMarginRtl:B,closeIconSize:M,closeBorderRadius:L,closeSize:F,closeMargin:j,closeMarginRtl:V,padding:N}=o,{type:a}=r,{left:O,right:q}=ce($);return{"--n-bezier":t,"--n-color":o[c("color",a)],"--n-close-icon-size":M,"--n-close-border-radius":L,"--n-close-color-hover":o[c("closeColorHover",a)],"--n-close-color-pressed":o[c("closeColorPressed",a)],"--n-close-icon-color":o[c("closeIconColor",a)],"--n-close-icon-color-hover":o[c("closeIconColorHover",a)],"--n-close-icon-color-pressed":o[c("closeIconColorPressed",a)],"--n-icon-color":o[c("iconColor",a)],"--n-border":o[c("border",a)],"--n-title-text-color":o[c("titleTextColor",a)],"--n-content-text-color":o[c("contentTextColor",a)],"--n-line-height":R,"--n-border-radius":P,"--n-font-size":S,"--n-title-font-weight":T,"--n-icon-size":w,"--n-icon-margin":$,"--n-icon-margin-rtl":B,"--n-close-size":F,"--n-close-margin":j,"--n-close-margin-rtl":V,"--n-padding":N,"--n-icon-margin-left":O,"--n-icon-margin-right":q}}),n=x?Q("alert",A(()=>r.type[0]),f,r):void 0,g=U(!0),b=()=>{const{onAfterLeave:t,onAfterHide:o}=r;t&&t(),o&&o()};return{rtlEnabled:k,mergedClsPrefix:e,mergedBordered:d,visible:g,handleCloseClick:()=>{var t;Promise.resolve((t=r.onClose)==null?void 0:t.call(r)).then(o=>{o!==!1&&(g.value=!1)})},handleAfterLeave:()=>{b()},mergedTheme:C,cssVars:x?void 0:f,themeClass:n==null?void 0:n.themeClass,onRender:n==null?void 0:n.onRender}},render(){var r;return(r=this.onRender)==null||r.call(this),s(),p(X,{onAfterLeave:this.handleAfterLeave},{default:()=>{const{mergedClsPrefix:e,$slots:d}=this,x={class:[`${e}-alert`,this.themeClass,this.closable&&`${e}-alert--closable`,this.showIcon&&`${e}-alert--show-icon`,!this.title&&this.closable&&`${e}-alert--right-adjust`,this.rtlEnabled&&`${e}-alert--rtl`],style:this.cssVars,role:"alert"};return this.visible?(s(),_("div",W({key:1},W(this.$attrs,x)),[v(()=>this.closable&&(s(),p(ee,{clsPrefix:e,class:y(`${e}-alert__close`),onClick:this.handleCloseClick},null,8,["clsPrefix","class","onClick"]))),v(()=>this.bordered&&(s(),_("div",{class:y(`${e}-alert__border`)},null,2))),v(()=>this.showIcon&&(s(),_("div",{class:y(`${e}-alert__icon`),"aria-hidden":"true"},[v(()=>oe(d.icon,()=>[(s(),p(re,{clsPrefix:e},{default:()=>{switch(this.type){case"success":return s(),p(se,{key:3});case"info":return s(),p(le,{key:4});case"warning":return s(),p(te,{key:5});case"error":return s(),p(ne,{key:6});default:return null}}},1032,["clsPrefix"]))]))],2))),ie("div",{class:y([`${e}-alert-body`,this.mergedBordered&&`${e}-alert-body--bordered`])},[v(()=>ae(d.header,l=>{const C=l||this.title;return C?(s(),_("div",{key:2,class:y(`${e}-alert-body__title`)},[v(()=>C)],2)):null})),v(()=>d.default&&(s(),_("div",{class:y(`${e}-alert-body__content`)},[v(()=>d.default())],2)))],2)],16)):null}},1032,["onAfterLeave"])}});export{Ce as A};
