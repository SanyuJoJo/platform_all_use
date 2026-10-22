import{a as W,S as Y,E as ee}from"./Spin-QgI_6sok.js";import{f as oe,ah as C,ai as b,i as S,j as f,R as L,aj as re,Q as te,d as N,k as F,o as l,s as h,c as I,n as m,h as A,l as k,L as M,ak as ne,p as se,X as le,q as ae,r as V,D as $,O as ie,al as ce,I as de,E as ue,W as fe,v as ge,S as ve,a1 as pe,a3 as he,t as v,Z as Ce,a as P,u as x,w as T,B as be,e as me,ab as j,ac as xe}from"./index-CVUczFfF.js";import{u as _e}from"./use-message-D3mBhG8n.js";import{I as Ie,p as ye}from"./proxy-BHDJblh3.js";import{_ as ze}from"./_plugin-vue_export-helper-DlAUqK2U.js";var Se={iconMargin:"11px 8px 0 12px",iconMarginRtl:"11px 12px 0 8px",iconSize:"24px",closeIconSize:"16px",closeSize:"20px",closeMargin:"13px 14px 0 0",closeMarginRtl:"13px 0 0 14px",padding:"13px"};function ke(n){const{lineHeight:e,borderRadius:i,fontWeightStrong:d,baseColor:s,dividerColor:c,actionColor:_,textColor1:u,textColor2:o,closeColorHover:t,closeColorPressed:p,closeIconColor:y,closeIconColorHover:z,closeIconColorPressed:a,infoColor:r,successColor:w,warningColor:E,errorColor:R,fontSize:B}=n;return{...Se,fontSize:B,lineHeight:e,titleFontWeight:d,borderRadius:i,border:`1px solid ${c}`,color:_,titleTextColor:u,iconColor:o,contentTextColor:o,closeBorderRadius:i,closeColorHover:t,closeColorPressed:p,closeIconColor:y,closeIconColorHover:z,closeIconColorPressed:a,borderInfo:`1px solid ${C(s,b(r,{alpha:.25}))}`,colorInfo:C(s,b(r,{alpha:.08})),titleTextColorInfo:u,iconColorInfo:r,contentTextColorInfo:o,closeColorHoverInfo:t,closeColorPressedInfo:p,closeIconColorInfo:y,closeIconColorHoverInfo:z,closeIconColorPressedInfo:a,borderSuccess:`1px solid ${C(s,b(w,{alpha:.25}))}`,colorSuccess:C(s,b(w,{alpha:.08})),titleTextColorSuccess:u,iconColorSuccess:w,contentTextColorSuccess:o,closeColorHoverSuccess:t,closeColorPressedSuccess:p,closeIconColorSuccess:y,closeIconColorHoverSuccess:z,closeIconColorPressedSuccess:a,borderWarning:`1px solid ${C(s,b(E,{alpha:.33}))}`,colorWarning:C(s,b(E,{alpha:.08})),titleTextColorWarning:u,iconColorWarning:E,contentTextColorWarning:o,closeColorHoverWarning:t,closeColorPressedWarning:p,closeIconColorWarning:y,closeIconColorHoverWarning:z,closeIconColorPressedWarning:a,borderError:`1px solid ${C(s,b(R,{alpha:.25}))}`,colorError:C(s,b(R,{alpha:.08})),titleTextColorError:u,iconColorError:R,contentTextColorError:o,closeColorHoverError:t,closeColorPressedError:p,closeIconColorError:y,closeIconColorHoverError:z,closeIconColorPressedError:a}}const Pe={common:oe,self:ke};var Te=S("alert",`
 line-height: var(--n-line-height);
 border-radius: var(--n-border-radius);
 position: relative;
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-color);
 text-align: start;
 word-break: break-word;
`,[f("border",`
 border-radius: inherit;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 transition: border-color .3s var(--n-bezier);
 border: var(--n-border);
 pointer-events: none;
 `),L("closable",[S("alert-body",[f("title",`
 padding-right: 24px;
 `)])]),f("icon",{color:"var(--n-icon-color)"}),S("alert-body",{padding:"var(--n-padding)"},[f("title",{color:"var(--n-title-text-color)"}),f("content",{color:"var(--n-content-text-color)"})]),re({originalTransition:"transform .3s var(--n-bezier)",enterToProps:{transform:"scale(1)"},leaveToProps:{transform:"scale(0.9)"}}),f("icon",`
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
 `),f("close",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 position: absolute;
 right: 0;
 top: 0;
 margin: var(--n-close-margin);
 `),L("show-icon",[S("alert-body",{paddingLeft:"calc(var(--n-icon-margin-left) + var(--n-icon-size) + var(--n-icon-margin-right))"})]),L("right-adjust",[S("alert-body",{paddingRight:"calc(var(--n-close-size) + var(--n-padding) + 2px)"})]),S("alert-body",`
 border-radius: var(--n-border-radius);
 transition: border-color .3s var(--n-bezier);
 `,[f("title",`
 transition: color .3s var(--n-bezier);
 font-size: 16px;
 line-height: 19px;
 font-weight: var(--n-title-font-weight);
 `,[te("& +",[f("content",{marginTop:"9px"})])]),f("content",{transition:"color .3s var(--n-bezier)",fontSize:"var(--n-font-size)"})]),f("icon",{transition:"color .3s var(--n-bezier)"})]);const we={...F.props,title:String,showIcon:{type:Boolean,default:!0},type:{type:String,default:"default"},bordered:{type:Boolean,default:!0},closable:Boolean,onClose:Function,onAfterLeave:Function,onAfterHide:Function};var Ee=N({name:"Alert",inheritAttrs:!1,props:we,slots:Object,setup(n){const{mergedClsPrefixRef:e,mergedBorderedRef:i,inlineThemeDisabled:d,mergedRtlRef:s}=se(n),c=F("Alert","-alert",Te,Pe,n,e),_=le("Alert",s,e),u=V(()=>{const{common:{cubicBezierEaseInOut:a},self:r}=c.value,{fontSize:w,borderRadius:E,titleFontWeight:R,lineHeight:B,iconSize:D,iconMargin:H,iconMarginRtl:O,closeIconSize:q,closeBorderRadius:K,closeSize:Q,closeMargin:U,closeMarginRtl:X,padding:Z}=r,{type:g}=n,{left:G,right:J}=he(H);return{"--n-bezier":a,"--n-color":r[v("color",g)],"--n-close-icon-size":q,"--n-close-border-radius":K,"--n-close-color-hover":r[v("closeColorHover",g)],"--n-close-color-pressed":r[v("closeColorPressed",g)],"--n-close-icon-color":r[v("closeIconColor",g)],"--n-close-icon-color-hover":r[v("closeIconColorHover",g)],"--n-close-icon-color-pressed":r[v("closeIconColorPressed",g)],"--n-icon-color":r[v("iconColor",g)],"--n-border":r[v("border",g)],"--n-title-text-color":r[v("titleTextColor",g)],"--n-content-text-color":r[v("contentTextColor",g)],"--n-line-height":B,"--n-border-radius":E,"--n-font-size":w,"--n-title-font-weight":R,"--n-icon-size":D,"--n-icon-margin":H,"--n-icon-margin-rtl":O,"--n-close-size":Q,"--n-close-margin":U,"--n-close-margin-rtl":X,"--n-padding":Z,"--n-icon-margin-left":G,"--n-icon-margin-right":J}}),o=d?ae("alert",V(()=>n.type[0]),u,n):void 0,t=$(!0),p=()=>{const{onAfterLeave:a,onAfterHide:r}=n;a&&a(),r&&r()};return{rtlEnabled:_,mergedClsPrefix:e,mergedBordered:i,visible:t,handleCloseClick:()=>{var a;Promise.resolve((a=n.onClose)==null?void 0:a.call(n)).then(r=>{r!==!1&&(t.value=!1)})},handleAfterLeave:()=>{p()},mergedTheme:c,cssVars:d?void 0:u,themeClass:o==null?void 0:o.themeClass,onRender:o==null?void 0:o.onRender}},render(){var n;return(n=this.onRender)==null||n.call(this),l(),h(ne,{onAfterLeave:this.handleAfterLeave},{default:()=>{const{mergedClsPrefix:e,$slots:i}=this,d={class:[`${e}-alert`,this.themeClass,this.closable&&`${e}-alert--closable`,this.showIcon&&`${e}-alert--show-icon`,!this.title&&this.closable&&`${e}-alert--right-adjust`,this.rtlEnabled&&`${e}-alert--rtl`],style:this.cssVars,role:"alert"};return this.visible?(l(),I("div",M({key:1},M(this.$attrs,d)),[m(()=>this.closable&&(l(),h(ie,{clsPrefix:e,class:k(`${e}-alert__close`),onClick:this.handleCloseClick},null,8,["clsPrefix","class","onClick"]))),m(()=>this.bordered&&(l(),I("div",{class:k(`${e}-alert__border`)},null,2))),m(()=>this.showIcon&&(l(),I("div",{class:k(`${e}-alert__icon`),"aria-hidden":"true"},[m(()=>ce(i.icon,()=>[(l(),h(de,{clsPrefix:e},{default:()=>{switch(this.type){case"success":return l(),h(ve,{key:3});case"info":return l(),h(ge,{key:4});case"warning":return l(),h(fe,{key:5});case"error":return l(),h(ue,{key:6});default:return null}}},1032,["clsPrefix"]))]))],2))),A("div",{class:k([`${e}-alert-body`,this.mergedBordered&&`${e}-alert-body--bordered`])},[m(()=>pe(i.header,s=>{const c=s||this.title;return c?(l(),I("div",{key:2,class:k(`${e}-alert-body__title`)},[m(()=>c)],2)):null})),m(()=>i.default&&(l(),I("div",{class:k(`${e}-alert-body__content`)},[m(()=>i.default())],2)))],2)],16)):null}},1032,["onAfterLeave"])}});const Re={class:"mono"},$e={key:0,class:"logs-tab__pre"},Ae=N({__name:"LogsTab",props:{serviceId:{}},setup(n){const e=n,i=_e(),d=$(!1),s=$(200),c=$(""),_=$("");async function u(){d.value=!0;try{const o=await ye.getLogs(e.serviceId,s.value);c.value=o.data.content||"",_.value=o.data.path||"",!c.value&&o.data.message&&i.info(o.data.message)}catch(o){i.error(o.message||"加载日志失败")}finally{d.value=!1}}return Ce(u),(o,t)=>(l(),I("div",null,[P(x(W),{justify:"space-between",align:"center",style:{"margin-bottom":"8px"}},{default:T(()=>[P(x(W),{align:"center"},{default:T(()=>[t[1]||(t[1]=A("span",{class:"logs-tab__title"},"日志",-1)),P(x(Ie),{value:s.value,"onUpdate:value":t[0]||(t[0]=p=>s.value=p),min:10,max:2e3,size:"small",style:{width:"120px"}},null,8,["value"]),t[2]||(t[2]=A("span",{class:"logs-tab__hint"},"最近 N 行",-1))]),_:1}),P(x(W),null,{default:T(()=>[P(x(be),{size:"small",loading:d.value,onClick:u},{default:T(()=>[...t[3]||(t[3]=[me("刷新",-1)])]),_:1},8,["loading"])]),_:1})]),_:1}),_.value?(l(),h(x(Ee),{key:0,type:"default","show-icon":!1,style:{"margin-bottom":"8px"}},{default:T(()=>[A("span",Re,j(_.value),1)]),_:1})):xe("",!0),P(x(Y),{show:d.value},{default:T(()=>[c.value?(l(),I("pre",$e,j(c.value),1)):(l(),h(x(ee),{key:1,description:"暂无日志（实例可能未启动）"}))]),_:1},8,["show"])]))}}),Ve=ze(Ae,[["__scopeId","data-v-9c552166"]]);export{Ve as default};
