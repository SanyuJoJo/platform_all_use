import{f as vt,ai as ut,g as Ke,h as P,Q as x,i as U,R as Y,j as N,aQ as ro,aR as lo,aS as ao,d as fe,k as be,a1 as Pt,o as b,c as M,n as Z,l as q,a as d,aM as no,ar as io,m as ce,J as Xe,p as Re,D as $,aC as ft,aV as Je,X as Qe,q as Ze,r as X,aT as so,a9 as G,t as De,a8 as ue,M as re,H as Ut,s as L,a7 as Fe,T as bt,U as co,F as uo,aw as fo,a4 as Ie,a5 as Oe,aB as Et,A as bo,aJ as Lt,L as qe,$ as ht,aW as vo,_ as ho,Y as Ge,ao as po,aX as mo,aY as pt,aZ as go,aG as _o,aD as yo,aF as ko,a_ as et,a$ as xo,b0 as wo,b1 as So,b2 as Co,b3 as zo,b4 as Ro,K as $o,O as To,aU as Bo,w as a,u as r,B as Te,e as S,ab as A,ac as te,b5 as Ve,ad as Fo,a2 as Ct}from"./index-CVUczFfF.js";import{T as se}from"./Tag-DPmVnj_2.js";import{u as Pe,p as de,a as je,I as zt}from"./proxy-BHDJblh3.js";import{f as Rt,F as at,l as ee,d as Be,m as Se,D as $t}from"./Switch-CwxL_nXb.js";import{g as Io,S as Do,a as oe}from"./Spin-QgI_6sok.js";import{a as nt,D as K}from"./DescriptionsItem-31qlb8JN.js";import{u as Po}from"./use-message-D3mBhG8n.js";import{C as it,L as Uo,b as Tt,d as Ce,e as Eo,S as Lo,p as st,i as ze,G as Bt,f as Ao,g as dt,D as Ft,h as It,j as Mo,k as Oo}from"./proxyOptions-BqbSonRo.js";import{_ as Ho}from"./_plugin-vue_export-helper-DlAUqK2U.js";var No={sizeSmall:"14px",sizeMedium:"16px",sizeLarge:"18px",labelPadding:"0 8px",labelFontWeight:"400"};function Vo(o){const{baseColor:l,inputColorDisabled:n,cardColor:v,modalColor:h,popoverColor:z,textColorDisabled:g,borderColor:y,primaryColor:p,textColor2:F,fontSizeSmall:I,fontSizeMedium:s,fontSizeLarge:m,borderRadiusSmall:k,lineHeight:C}=o;return{...No,labelLineHeight:C,fontSizeSmall:I,fontSizeMedium:s,fontSizeLarge:m,borderRadius:k,color:l,colorChecked:p,colorDisabled:n,colorDisabledChecked:n,colorTableHeader:v,colorTableHeaderModal:h,colorTableHeaderPopover:z,checkMarkColor:l,checkMarkColorDisabled:g,checkMarkColorDisabledChecked:g,border:`1px solid ${y}`,borderDisabled:`1px solid ${y}`,borderDisabledChecked:`1px solid ${y}`,borderChecked:`1px solid ${p}`,borderFocus:`1px solid ${p}`,boxShadowFocus:`0 0 0 2px ${ut(p,{alpha:.3})}`,textColor:F,textColorDisabled:g}}const jo={name:"Checkbox",common:vt,self:Vo};var Wo=()=>(()=>{const o=Ke("75be776d8875fa17");return o[0]||(o[0]=P("svg",{viewBox:"0 0 64 64",class:"check-icon"},[P("path",{d:"M50.42,16.76L22.34,39.45l-8.1-11.46c-1.12-1.58-3.3-1.96-4.88-0.84c-1.58,1.12-1.95,3.3-0.84,4.88l10.26,14.51  c0.56,0.79,1.42,1.31,2.38,1.45c0.16,0.02,0.32,0.03,0.48,0.03c0.8,0,1.57-0.27,2.2-0.78l30.99-25.03c1.5-1.21,1.74-3.42,0.52-4.92  C54.13,15.78,51.93,15.55,50.42,16.76z"})],-1))})(),Go=()=>(()=>{const o=Ke("c6eed899356c8404");return o[0]||(o[0]=P("svg",{viewBox:"0 0 100 100",class:"line-icon"},[P("path",{d:"M80.2,55.5H21.4c-2.8,0-5.1-2.5-5.1-5.5l0,0c0-3,2.3-5.5,5.1-5.5h58.7c2.8,0,5.1,2.5,5.1,5.5l0,0C85.2,53.1,82.9,55.5,80.2,55.5z"})],-1))})(),Yo=x([U("checkbox",`
 font-size: var(--n-font-size);
 outline: none;
 cursor: pointer;
 display: inline-flex;
 flex-wrap: nowrap;
 align-items: flex-start;
 word-break: break-word;
 line-height: var(--n-size);
 --n-merged-color-table: var(--n-color-table);
 `,[Y("show-label","line-height: var(--n-label-line-height);"),x("&:hover",[U("checkbox-box",[N("border","border: var(--n-border-checked);")])]),x("&:focus:not(:active)",[U("checkbox-box",[N("border",`
 border: var(--n-border-focus);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),Y("inside-table",[U("checkbox-box",`
 background-color: var(--n-merged-color-table);
 `)]),Y("checked",[U("checkbox-box",`
 background-color: var(--n-color-checked);
 `,[U("checkbox-icon",[x(".check-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),Y("indeterminate",[U("checkbox-box",[U("checkbox-icon",[x(".check-icon",`
 opacity: 0;
 transform: scale(.5);
 `),x(".line-icon",`
 opacity: 1;
 transform: scale(1);
 `)])])]),Y("checked, indeterminate",[x("&:focus:not(:active)",[U("checkbox-box",[N("border",`
 border: var(--n-border-checked);
 box-shadow: var(--n-box-shadow-focus);
 `)])]),U("checkbox-box",`
 background-color: var(--n-color-checked);
 border-left: 0;
 border-top: 0;
 `,[N("border",{border:"var(--n-border-checked)"})])]),Y("disabled",{cursor:"not-allowed"},[Y("checked",[U("checkbox-box",`
 background-color: var(--n-color-disabled-checked);
 `,[N("border",{border:"var(--n-border-disabled-checked)"}),U("checkbox-icon",[x(".check-icon, .line-icon",{fill:"var(--n-check-mark-color-disabled-checked)"})])])]),U("checkbox-box",`
 background-color: var(--n-color-disabled);
 `,[N("border",`
 border: var(--n-border-disabled);
 `),U("checkbox-icon",[x(".check-icon, .line-icon",`
 fill: var(--n-check-mark-color-disabled);
 `)])]),N("label",`
 color: var(--n-text-color-disabled);
 `)]),U("checkbox-box-wrapper",`
 position: relative;
 width: var(--n-size);
 flex-shrink: 0;
 flex-grow: 0;
 user-select: none;
 -webkit-user-select: none;
 `),U("checkbox-box",`
 position: absolute;
 left: 0;
 top: 50%;
 transform: translateY(-50%);
 height: var(--n-size);
 width: var(--n-size);
 display: inline-block;
 box-sizing: border-box;
 border-radius: var(--n-border-radius);
 background-color: var(--n-color);
 transition: background-color 0.3s var(--n-bezier);
 `,[N("border",`
 transition:
 border-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 border-radius: inherit;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 border: var(--n-border);
 `),U("checkbox-icon",`
 display: flex;
 align-items: center;
 justify-content: center;
 position: absolute;
 left: 1px;
 right: 1px;
 top: 1px;
 bottom: 1px;
 `,[x(".check-icon, .line-icon",`
 width: 100%;
 fill: var(--n-check-mark-color);
 opacity: 0;
 transform: scale(0.5);
 transform-origin: center;
 transition:
 fill 0.3s var(--n-bezier),
 transform 0.3s var(--n-bezier),
 opacity 0.3s var(--n-bezier),
 border-color 0.3s var(--n-bezier);
 `),ro({left:"1px",top:"1px"})])]),N("label",`
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 user-select: none;
 -webkit-user-select: none;
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 `,[x("&:empty",{display:"none"})])]),lo(U("checkbox",`
 --n-merged-color-table: var(--n-color-table-modal);
 `)),ao(U("checkbox",`
 --n-merged-color-table: var(--n-color-table-popover);
 `))]);const qo=["id"],Ko=["tabindex","aria-checked","aria-labelledby","onKeyup","onKeydown","onClick"],Xo={...be.props,size:String,checked:{type:[Boolean,String,Number],default:void 0},defaultChecked:{type:[Boolean,String,Number],default:!1},value:[String,Number],disabled:{type:Boolean,default:void 0},indeterminate:Boolean,label:String,focusable:{type:Boolean,default:!0},checkedValue:{type:[Boolean,String,Number],default:!0},uncheckedValue:{type:[Boolean,String,Number],default:!1},"onUpdate:checked":[Function,Array],onUpdateChecked:[Function,Array],privateInsideTable:Boolean,onChange:[Function,Array]};var Me=fe({name:"Checkbox",props:Xo,setup(o){const l=Xe(At,null),n=$(null),{mergedClsPrefixRef:v,inlineThemeDisabled:h,mergedRtlRef:z,mergedComponentPropsRef:g}=Re(o),y=$(o.defaultChecked),p=ue(o,"checked"),F=Pe(p,y),I=ft(()=>{if(l){const f=l.valueSetRef.value;return f&&o.value!==void 0?f.has(o.value):!1}else return F.value===o.checkedValue}),s=Je(o,{mergedSize(f){var E,H;const{size:D}=o;if(D!==void 0)return D;if(l){const{value:j}=l.mergedSizeRef;if(j!==void 0)return j}if(f){const{mergedSize:j}=f;if(j!==void 0)return j.value}const _=(H=(E=g==null?void 0:g.value)==null?void 0:E.Checkbox)==null?void 0:H.size;return _||"medium"},mergedDisabled(f){const{disabled:D}=o;if(D!==void 0)return D;if(l){if(l.disabledRef.value)return!0;const{maxRef:{value:_},checkedCountRef:E}=l;if(_!==void 0&&E.value>=_&&!I.value)return!0;const{minRef:{value:H}}=l;if(H!==void 0&&E.value<=H&&I.value)return!0}return f?f.disabled.value:!1}}),{mergedDisabledRef:m,mergedSizeRef:k}=s,C=be("Checkbox","-checkbox",Yo,jo,o,v);function T(f){if(l&&o.value!==void 0)l.toggleCheckbox(!I.value,o.value);else{const{onChange:D,"onUpdate:checked":_,onUpdateChecked:E}=o,{nTriggerFormInput:H,nTriggerFormChange:j}=s,Q=I.value?o.uncheckedValue:o.checkedValue;_&&G(_,Q,f),E&&G(E,Q,f),D&&G(D,Q,f),H(),j(),y.value=Q}}function V(f){m.value||T(f)}function W(f){if(!m.value)switch(f.key){case" ":case"Enter":T(f)}}function R(f){switch(f.key){case" ":f.preventDefault()}}const O={focus:()=>{var f;(f=n.value)==null||f.focus()},blur:()=>{var f;(f=n.value)==null||f.blur()}},B=Qe("Checkbox",z,v),w=X(()=>{const{value:f}=k,{common:{cubicBezierEaseInOut:D},self:{borderRadius:_,color:E,colorChecked:H,colorDisabled:j,colorTableHeader:Q,colorTableHeaderModal:le,colorTableHeaderPopover:ne,checkMarkColor:ve,checkMarkColorDisabled:he,border:pe,borderFocus:me,borderDisabled:ie,borderChecked:ge,boxShadowFocus:_e,textColor:ye,textColorDisabled:ke,checkMarkColorDisabledChecked:xe,colorDisabledChecked:Ue,borderDisabledChecked:c,labelPadding:Ee,labelLineHeight:$e,labelFontWeight:we,[De("fontSize",f)]:tt,[De("size",f)]:ot}}=C.value;return{"--n-label-line-height":$e,"--n-label-font-weight":we,"--n-size":ot,"--n-bezier":D,"--n-border-radius":_,"--n-border":pe,"--n-border-checked":ge,"--n-border-focus":me,"--n-border-disabled":ie,"--n-border-disabled-checked":c,"--n-box-shadow-focus":_e,"--n-color":E,"--n-color-checked":H,"--n-color-table":Q,"--n-color-table-modal":le,"--n-color-table-popover":ne,"--n-color-disabled":j,"--n-color-disabled-checked":Ue,"--n-text-color":ye,"--n-text-color-disabled":ke,"--n-check-mark-color":ve,"--n-check-mark-color-disabled":he,"--n-check-mark-color-disabled-checked":xe,"--n-font-size":tt,"--n-label-padding":Ee}}),u=h?Ze("checkbox",X(()=>k.value[0]),w,o):void 0;return Object.assign(s,O,{rtlEnabled:B,selfRef:n,mergedClsPrefix:v,mergedDisabled:m,renderedChecked:I,mergedTheme:C,labelId:so(),handleClick:V,handleKeyUp:W,handleKeyDown:R,cssVars:h?void 0:w,themeClass:u==null?void 0:u.themeClass,onRender:u==null?void 0:u.onRender})},render(){var C;const{$slots:o,renderedChecked:l,mergedDisabled:n,indeterminate:v,privateInsideTable:h,cssVars:z,labelId:g,label:y,mergedClsPrefix:p,focusable:F,handleKeyUp:I,handleKeyDown:s,handleClick:m}=this;(C=this.onRender)==null||C.call(this);const k=Pt(o.default,T=>y||T?(b(),M("span",{key:1,class:q(`${p}-checkbox__label`),id:g},[Z(()=>y||T)],10,qo)):null);return(()=>{const T=Ke("70be6e74cd27cb50");return b(),M("div",{ref:"selfRef",class:q([`${p}-checkbox`,this.themeClass,this.rtlEnabled&&`${p}-checkbox--rtl`,l&&`${p}-checkbox--checked`,n&&`${p}-checkbox--disabled`,v&&`${p}-checkbox--indeterminate`,h&&`${p}-checkbox--inside-table`,k&&`${p}-checkbox--show-label`]),tabindex:n||!F?void 0:0,role:"checkbox","aria-checked":v?"mixed":l,"aria-labelledby":g,style:ce(z),onKeyup:I,onKeydown:s,onClick:m,onMousedown:T[0]||(T[0]=()=>{io("selectstart",window,V=>{V.preventDefault()},{once:!0})})},[P("div",{class:q(`${p}-checkbox-box-wrapper`)},[T[1]||(T[1]=Z(" ",-1)),P("div",{class:q(`${p}-checkbox-box`)},[d(no,null,{default:()=>this.indeterminate?(b(),M("div",{key:"indeterminate",class:q(`${p}-checkbox-icon`)},[Z(()=>Go())],2)):(b(),M("div",{key:"check",class:q(`${p}-checkbox-icon`)},[Z(()=>Wo())],2))},1024),P("div",{class:q(`${p}-checkbox-box__border`)},null,2)],2)],2),Z(()=>k)],46,Ko)})()}});const At=Ut("n-checkbox-group"),Jo={min:Number,max:Number,size:String,options:Array,labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},value:Array,defaultValue:{type:Array,default:null},disabled:{type:Boolean,default:void 0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],onChange:[Function,Array]};var We=fe({name:"CheckboxGroup",props:Jo,setup(o){const{mergedClsPrefixRef:l}=Re(o),n=Je(o),{mergedSizeRef:v,mergedDisabledRef:h}=n,z=$(o.defaultValue),g=X(()=>o.value),y=Pe(g,z),p=X(()=>{var s;return((s=y.value)==null?void 0:s.length)||0}),F=X(()=>Array.isArray(y.value)?new Set(y.value):new Set);function I(s,m){const{nTriggerFormInput:k,nTriggerFormChange:C}=n,{onChange:T,"onUpdate:value":V,onUpdateValue:W}=o;if(Array.isArray(y.value)){const R=Array.from(y.value),O=R.findIndex(B=>B===m);s?~O||(R.push(m),W&&G(W,R,{actionType:"check",value:m}),V&&G(V,R,{actionType:"check",value:m}),k(),C(),z.value=R,T&&G(T,R)):~O&&(R.splice(O,1),W&&G(W,R,{actionType:"uncheck",value:m}),V&&G(V,R,{actionType:"uncheck",value:m}),T&&G(T,R),z.value=R,k(),C())}else s?(W&&G(W,[m],{actionType:"check",value:m}),V&&G(V,[m],{actionType:"check",value:m}),T&&G(T,[m]),z.value=[m],k(),C()):(W&&G(W,[],{actionType:"uncheck",value:m}),V&&G(V,[],{actionType:"uncheck",value:m}),T&&G(T,[]),z.value=[],k(),C())}return Fe(At,{checkedCountRef:p,maxRef:ue(o,"max"),minRef:ue(o,"min"),valueSetRef:F,disabledRef:h,mergedSizeRef:v,toggleCheckbox:I}),{mergedClsPrefix:l}},render(){const{options:o,labelField:l,valueField:n}=this.$props;return b(),M("div",{class:q(`${this.mergedClsPrefix}-checkbox-group`),role:"group"},[o?(b(),M(re,{key:0},[Z(()=>o.map(v=>{const h=v[n];return b(),L(Me,{key:h,value:h,disabled:v.disabled,label:v[l]},null,8,["value","disabled","label"])}))],64)):(b(),M(re,{key:1},[Z(()=>{var v,h;return(h=(v=this.$slots).default)==null?void 0:h.call(v)})],64))],2)}}),Qo={radioSizeSmall:"14px",radioSizeMedium:"16px",radioSizeLarge:"18px",labelPadding:"0 8px",labelFontWeight:"400"};function Zo(o){const{borderColor:l,primaryColor:n,baseColor:v,textColorDisabled:h,inputColorDisabled:z,textColor2:g,opacityDisabled:y,borderRadius:p,fontSizeSmall:F,fontSizeMedium:I,fontSizeLarge:s,heightSmall:m,heightMedium:k,heightLarge:C,lineHeight:T}=o;return{...Qo,labelLineHeight:T,buttonHeightSmall:m,buttonHeightMedium:k,buttonHeightLarge:C,fontSizeSmall:F,fontSizeMedium:I,fontSizeLarge:s,boxShadow:`inset 0 0 0 1px ${l}`,boxShadowActive:`inset 0 0 0 1px ${n}`,boxShadowFocus:`inset 0 0 0 1px ${n}, 0 0 0 2px ${ut(n,{alpha:.2})}`,boxShadowHover:`inset 0 0 0 1px ${n}`,boxShadowDisabled:`inset 0 0 0 1px ${l}`,color:v,colorDisabled:z,colorActive:"#0000",textColor:g,textColorDisabled:h,dotColorActive:n,dotColorDisabled:l,buttonBorderColor:l,buttonBorderColorActive:n,buttonBorderColorHover:l,buttonColor:v,buttonColorActive:v,buttonTextColor:g,buttonTextColorActive:n,buttonTextColorHover:n,opacityDisabled:y,buttonBoxShadowFocus:`inset 0 0 0 1px ${n}, 0 0 0 2px ${ut(n,{alpha:.3})}`,buttonBoxShadowHover:"inset 0 0 0 1px #0000",buttonBoxShadow:"inset 0 0 0 1px #0000",buttonBorderRadius:p}}const Mt={name:"Radio",common:vt,self:Zo};var er=U("radio",`
 line-height: var(--n-label-line-height);
 outline: none;
 position: relative;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 align-items: flex-start;
 flex-wrap: nowrap;
 font-size: var(--n-font-size);
 word-break: break-word;
`,[Y("checked",[N("dot",`
 background-color: var(--n-color-active);
 `)]),N("dot-wrapper",`
 position: relative;
 flex-shrink: 0;
 flex-grow: 0;
 width: var(--n-radio-size);
 `),U("radio-input",`
 position: absolute;
 border: 0;
 width: 0;
 height: 0;
 opacity: 0;
 margin: 0;
 `),N("dot",`
 position: absolute;
 top: 50%;
 left: 0;
 transform: translateY(-50%);
 height: var(--n-radio-size);
 width: var(--n-radio-size);
 background: var(--n-color);
 box-shadow: var(--n-box-shadow);
 border-radius: 50%;
 transition:
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `,[x("&::before",`
 content: "";
 opacity: 0;
 position: absolute;
 left: 4px;
 top: 4px;
 height: calc(100% - 8px);
 width: calc(100% - 8px);
 border-radius: 50%;
 transform: scale(.8);
 background: var(--n-dot-color-active);
 transition: 
 opacity .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),Y("checked",{boxShadow:"var(--n-box-shadow-active)"},[x("&::before",`
 opacity: 1;
 transform: scale(1);
 `)])]),N("label",`
 color: var(--n-text-color);
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 display: inline-block;
 transition: color .3s var(--n-bezier);
 `),bt("disabled",`
 cursor: pointer;
 `,[x("&:hover",[N("dot",{boxShadow:"var(--n-box-shadow-hover)"})]),Y("focus",[x("&:not(:active)",[N("dot",{boxShadow:"var(--n-box-shadow-focus)"})])])]),Y("disabled",`
 cursor: not-allowed;
 `,[N("dot",{boxShadow:"var(--n-box-shadow-disabled)",backgroundColor:"var(--n-color-disabled)"},[x("&::before",{backgroundColor:"var(--n-dot-color-disabled)"}),Y("checked",`
 opacity: 1;
 `)]),N("label",{color:"var(--n-text-color-disabled)"}),U("radio-input",`
 cursor: not-allowed;
 `)])]);const tr={name:String,value:{type:[String,Number,Boolean],default:"on"},checked:{type:Boolean,default:void 0},defaultChecked:Boolean,disabled:{type:Boolean,default:void 0},label:String,size:String,onUpdateChecked:[Function,Array],"onUpdate:checked":[Function,Array],checkedValue:{type:Boolean,default:void 0}},Ot=Ut("n-radio-group");function or(o){const l=Xe(Ot,null),{mergedClsPrefixRef:n,mergedComponentPropsRef:v}=Re(o),h=Je(o,{mergedSize(B){var f,D;const{size:w}=o;if(w!==void 0)return w;if(l){const{mergedSizeRef:{value:_}}=l;if(_!==void 0)return _}if(B)return B.mergedSize.value;const u=(D=(f=v==null?void 0:v.value)==null?void 0:f.Radio)==null?void 0:D.size;return u||"medium"},mergedDisabled(B){return!!(o.disabled||l!=null&&l.disabledRef.value||B!=null&&B.disabled.value)}}),{mergedSizeRef:z,mergedDisabledRef:g}=h,y=$(null),p=$(null),F=$(o.defaultChecked),I=ue(o,"checked"),s=Pe(I,F),m=ft(()=>l?l.valueRef.value===o.value:s.value),k=ft(()=>{const{name:B}=o;if(B!==void 0)return B;if(l)return l.nameRef.value}),C=$(!1);function T(){if(l){const{doUpdateValue:B}=l,{value:w}=o;G(B,w)}else{const{onUpdateChecked:B,"onUpdate:checked":w}=o,{nTriggerFormInput:u,nTriggerFormChange:f}=h;B&&G(B,!0),w&&G(w,!0),u(),f(),F.value=!0}}function V(){g.value||m.value||T()}function W(){V(),y.value&&(y.value.checked=m.value)}function R(){C.value=!1}function O(){C.value=!0}return{mergedClsPrefix:l?l.mergedClsPrefixRef:n,inputRef:y,labelRef:p,mergedName:k,mergedDisabled:g,renderSafeChecked:m,focus:C,mergedSize:z,handleRadioInputChange:W,handleRadioInputBlur:R,handleRadioInputFocus:O}}const rr=["value","name","checked","disabled","onChange","onFocus","onBlur"],lr={...be.props,...tr};var Ye=fe({name:"Radio",props:lr,setup(o){const l=or(o),n=be("Radio","-radio",er,Mt,o,l.mergedClsPrefix),v=X(()=>{const{mergedSize:{value:F}}=l,{common:{cubicBezierEaseInOut:I},self:{boxShadow:s,boxShadowActive:m,boxShadowDisabled:k,boxShadowFocus:C,boxShadowHover:T,color:V,colorDisabled:W,colorActive:R,textColor:O,textColorDisabled:B,dotColorActive:w,dotColorDisabled:u,labelPadding:f,labelLineHeight:D,labelFontWeight:_,[De("fontSize",F)]:E,[De("radioSize",F)]:H}}=n.value;return{"--n-bezier":I,"--n-label-line-height":D,"--n-label-font-weight":_,"--n-box-shadow":s,"--n-box-shadow-active":m,"--n-box-shadow-disabled":k,"--n-box-shadow-focus":C,"--n-box-shadow-hover":T,"--n-color":V,"--n-color-active":R,"--n-color-disabled":W,"--n-dot-color-active":w,"--n-dot-color-disabled":u,"--n-font-size":E,"--n-radio-size":H,"--n-text-color":O,"--n-text-color-disabled":B,"--n-label-padding":f}}),{inlineThemeDisabled:h,mergedClsPrefixRef:z,mergedRtlRef:g}=Re(o),y=Qe("Radio",g,z),p=h?Ze("radio",X(()=>l.mergedSize.value[0]),v,o):void 0;return Object.assign(l,{rtlEnabled:y,cssVars:h?void 0:v,themeClass:p==null?void 0:p.themeClass,onRender:p==null?void 0:p.onRender})},render(){const{$slots:o,mergedClsPrefix:l,onRender:n,label:v}=this;return n==null||n(),(()=>{const h=Ke("f8c6901d8cd45c02");return b(),M("label",{class:q([`${l}-radio`,this.themeClass,this.rtlEnabled&&`${l}-radio--rtl`,this.mergedDisabled&&`${l}-radio--disabled`,this.renderSafeChecked&&`${l}-radio--checked`,this.focus&&`${l}-radio--focus`]),style:ce(this.cssVars)},[P("div",{class:q(`${l}-radio__dot-wrapper`)},[h[0]||(h[0]=Z(" ",-1)),P("div",{class:q([`${l}-radio__dot`,this.renderSafeChecked&&`${l}-radio__dot--checked`])},null,2),P("input",{ref:"inputRef",type:"radio",class:q(`${l}-radio-input`),value:this.value,name:this.mergedName,checked:this.renderSafeChecked,disabled:this.mergedDisabled,onChange:this.handleRadioInputChange,onFocus:this.handleRadioInputFocus,onBlur:this.handleRadioInputBlur},null,42,rr)],2),Z(()=>Pt(o.default,z=>!z&&!v?null:(b(),M("div",{ref:"labelRef",class:q(`${l}-radio__label`)},[Z(()=>z||v)],2))))],6)})()}}),ar=U("radio-group",`
 display: inline-block;
 font-size: var(--n-font-size);
`,[N("splitor",`
 display: inline-block;
 vertical-align: bottom;
 width: 1px;
 transition:
 background-color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 background: var(--n-button-border-color);
 `,[Y("checked",{backgroundColor:"var(--n-button-border-color-active)"}),Y("disabled",{opacity:"var(--n-opacity-disabled)"})]),Y("button-group",`
 white-space: nowrap;
 height: var(--n-height);
 line-height: var(--n-height);
 `,[U("radio-button",{height:"var(--n-height)",lineHeight:"var(--n-height)"}),N("splitor",{height:"var(--n-height)"})]),U("radio-button",`
 vertical-align: bottom;
 outline: none;
 position: relative;
 user-select: none;
 -webkit-user-select: none;
 display: inline-block;
 box-sizing: border-box;
 padding-left: 14px;
 padding-right: 14px;
 white-space: nowrap;
 transition:
 background-color .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background: var(--n-button-color);
 color: var(--n-button-text-color);
 border-top: 1px solid var(--n-button-border-color);
 border-bottom: 1px solid var(--n-button-border-color);
 `,[U("radio-input",`
 pointer-events: none;
 position: absolute;
 border: 0;
 border-radius: inherit;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 opacity: 0;
 z-index: 1;
 `),N("state-border",`
 z-index: 1;
 pointer-events: none;
 position: absolute;
 box-shadow: var(--n-button-box-shadow);
 transition: box-shadow .3s var(--n-bezier);
 left: -1px;
 bottom: -1px;
 right: -1px;
 top: -1px;
 `),x("&:first-child",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 border-left: 1px solid var(--n-button-border-color);
 `,[N("state-border",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 `)]),x("&:last-child",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 border-right: 1px solid var(--n-button-border-color);
 `,[N("state-border",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 `)]),bt("disabled",`
 cursor: pointer;
 `,[x("&:hover",[N("state-border",`
 transition: box-shadow .3s var(--n-bezier);
 box-shadow: var(--n-button-box-shadow-hover);
 `),bt("checked",{color:"var(--n-button-text-color-hover)"})]),Y("focus",[x("&:not(:active)",[N("state-border",{boxShadow:"var(--n-button-box-shadow-focus)"})])])]),Y("checked",`
 background: var(--n-button-color-active);
 color: var(--n-button-text-color-active);
 border-color: var(--n-button-border-color-active);
 `),Y("disabled",`
 cursor: not-allowed;
 opacity: var(--n-opacity-disabled);
 `)])]);const nr=["onFocusin","onFocusout"];function ir(o,l,n){var z;const v=[];let h=!1;for(let g=0;g<o.length;++g){const y=o[g],p=(z=y.type)==null?void 0:z.name;p==="RadioButton"&&(h=!0);const F=y.props;if(p!=="RadioButton"){v.push(y);continue}if(g===0)v.push(y);else{const I=v[v.length-1].props,s=l===I.value,m=I.disabled,k=l===F.value,C=F.disabled,T=(s?2:0)+(m?0:1),V=(k?2:0)+(C?0:1),W={[`${n}-radio-group__splitor--disabled`]:m,[`${n}-radio-group__splitor--checked`]:s},R={[`${n}-radio-group__splitor--disabled`]:C,[`${n}-radio-group__splitor--checked`]:k},O=T<V?R:W;v.push((b(),M("div",{key:1,class:q([`${n}-radio-group__splitor`,O])},null,2)),y)}}return{children:v,isButtonGroup:h}}const sr={...be.props,name:String,options:Array,labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},value:[String,Number,Boolean],defaultValue:{type:[String,Number,Boolean],default:null},size:String,disabled:{type:Boolean,default:void 0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array]};var dr=fe({name:"RadioGroup",props:sr,setup(o){const l=$(null),{mergedSizeRef:n,mergedDisabledRef:v,nTriggerFormChange:h,nTriggerFormInput:z,nTriggerFormBlur:g,nTriggerFormFocus:y}=Je(o),{mergedClsPrefixRef:p,inlineThemeDisabled:F,mergedRtlRef:I}=Re(o),s=be("Radio","-radio-group",ar,Mt,o,p),m=$(o.defaultValue),k=ue(o,"value"),C=Pe(k,m);function T(w){const{onUpdateValue:u,"onUpdate:value":f}=o;u&&G(u,w),f&&G(f,w),m.value=w,h(),z()}function V(w){const{value:u}=l;u&&(u.contains(w.relatedTarget)||y())}function W(w){const{value:u}=l;u&&(u.contains(w.relatedTarget)||g())}Fe(Ot,{mergedClsPrefixRef:p,nameRef:ue(o,"name"),valueRef:C,disabledRef:v,mergedSizeRef:n,doUpdateValue:T});const R=Qe("Radio",I,p),O=X(()=>{const{value:w}=n,{common:{cubicBezierEaseInOut:u},self:{buttonBorderColor:f,buttonBorderColorActive:D,buttonBorderRadius:_,buttonBoxShadow:E,buttonBoxShadowFocus:H,buttonBoxShadowHover:j,buttonColor:Q,buttonColorActive:le,buttonTextColor:ne,buttonTextColorActive:ve,buttonTextColorHover:he,opacityDisabled:pe,[De("buttonHeight",w)]:me,[De("fontSize",w)]:ie}}=s.value;return{"--n-font-size":ie,"--n-bezier":u,"--n-button-border-color":f,"--n-button-border-color-active":D,"--n-button-border-radius":_,"--n-button-box-shadow":E,"--n-button-box-shadow-focus":H,"--n-button-box-shadow-hover":j,"--n-button-color":Q,"--n-button-color-active":le,"--n-button-text-color":ne,"--n-button-text-color-hover":he,"--n-button-text-color-active":ve,"--n-height":me,"--n-opacity-disabled":pe}}),B=F?Ze("radio-group",X(()=>n.value[0]),O,o):void 0;return{selfElRef:l,rtlEnabled:R,mergedClsPrefix:p,mergedValue:C,handleFocusout:W,handleFocusin:V,cssVars:F?void 0:O,themeClass:B==null?void 0:B.themeClass,onRender:B==null?void 0:B.onRender}},render(){var F;const{mergedValue:o,mergedClsPrefix:l,handleFocusin:n,handleFocusout:v}=this,{options:h,labelField:z,valueField:g}=this.$props,{children:y,isButtonGroup:p}=ir(h?h.map(I=>{const s=I[g];return b(),L(Ye,{key:typeof s=="boolean"?`__n_${s}`:s,value:s,disabled:I.disabled,label:I[z]},null,8,["value","disabled","label"])}):co(Io(this)),o,l);return(F=this.onRender)==null||F.call(this),b(),M("div",{onFocusin:n,onFocusout:v,ref:"selfElRef",class:q([`${l}-radio-group`,this.rtlEnabled&&`${l}-radio-group--rtl`,this.themeClass,p&&`${l}-radio-group--button-group`]),style:ce(this.cssVars)},[Z(()=>y)],46,nr)}});function cr(o){const{modalColor:l,textColor1:n,textColor2:v,boxShadow3:h,lineHeight:z,fontWeightStrong:g,dividerColor:y,closeColorHover:p,closeColorPressed:F,closeIconColor:I,closeIconColorHover:s,closeIconColorPressed:m,borderRadius:k,primaryColorHover:C}=o;return{bodyPadding:"16px 24px",borderRadius:k,headerPadding:"16px 24px",footerPadding:"16px 24px",color:l,textColor:v,titleTextColor:n,titleFontSize:"18px",titleFontWeight:g,boxShadow:h,lineHeight:z,headerBorderBottom:`1px solid ${y}`,footerBorderTop:`1px solid ${y}`,closeIconColor:I,closeIconColorHover:s,closeIconColorPressed:m,closeSize:"22px",closeIconSize:"18px",closeColorHover:p,closeColorPressed:F,closeBorderRadius:k,resizableTriggerColorHover:C}}const ur=uo({name:"Drawer",common:vt,peers:{Scrollbar:fo},self:cr}),fr=["onMouseenter","onMouseleave","onMousedown"],br={key:1,role:"none"};var vr=fe({name:"NDrawerContent",inheritAttrs:!1,props:{blockScroll:Boolean,show:{type:Boolean,default:void 0},displayDirective:{type:String,required:!0},placement:{type:String,required:!0},contentClass:String,contentStyle:[Object,String],nativeScrollbar:{type:Boolean,required:!0},scrollbarProps:Object,trapFocus:{type:Boolean,default:!0},autoFocus:{type:Boolean,default:!0},showMask:{type:[Boolean,String],required:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,onClickoutside:Function,onAfterLeave:Function,onAfterEnter:Function,onEsc:Function},setup(o){const l=$(!!o.show),n=$(null),v=Xe(pt);let h=0,z="",g=null;const y=$(!1),p=$(!1),F=X(()=>o.placement==="top"||o.placement==="bottom"),{mergedClsPrefixRef:I,mergedRtlRef:s}=Re(o),m=Qe("Drawer",s,I),k=u,C=_=>{p.value=!0,h=F.value?_.clientY:_.clientX,z=document.body.style.cursor,document.body.style.cursor=F.value?"ns-resize":"ew-resize",document.body.addEventListener("mousemove",w),document.body.addEventListener("mouseleave",k),document.body.addEventListener("mouseup",u)},T=()=>{g!==null&&(window.clearTimeout(g),g=null),p.value?y.value=!0:g=window.setTimeout(()=>{y.value=!0},300)},V=()=>{g!==null&&(window.clearTimeout(g),g=null),y.value=!1},{doUpdateHeight:W,doUpdateWidth:R}=v,O=_=>{const{maxWidth:E}=o;if(E&&_>E)return E;const{minWidth:H}=o;return H&&_<H?H:_},B=_=>{const{maxHeight:E}=o;if(E&&_>E)return E;const{minHeight:H}=o;return H&&_<H?H:_};function w(_){var E,H;if(p.value)if(F.value){let j=((E=n.value)==null?void 0:E.offsetHeight)||0;const Q=h-_.clientY;j+=o.placement==="bottom"?Q:-Q,j=B(j),W(j),h=_.clientY}else{let j=((H=n.value)==null?void 0:H.offsetWidth)||0;const Q=h-_.clientX;j+=o.placement==="right"?Q:-Q,j=O(j),R(j),h=_.clientX}}function u(){p.value&&(h=0,p.value=!1,document.body.style.cursor=z,document.body.removeEventListener("mousemove",w),document.body.removeEventListener("mouseup",u),document.body.removeEventListener("mouseleave",k))}ho(()=>{o.show&&(l.value=!0)}),Ge(()=>o.show,_=>{_||u()}),po(()=>{u()});const f=X(()=>{const{show:_}=o,E=[[Oe,_]];return o.showMask||E.push([go,o.onClickoutside,void 0,{capture:!0}]),E});function D(){var _;l.value=!1,(_=o.onAfterLeave)==null||_.call(o)}return mo(X(()=>o.blockScroll&&l.value)),Fe(_o,n),Fe(yo,null),Fe(ko,null),{bodyRef:n,rtlEnabled:m,mergedClsPrefix:v.mergedClsPrefixRef,isMounted:v.isMountedRef,mergedTheme:v.mergedThemeRef,displayed:l,transitionName:X(()=>({right:"slide-in-from-right-transition",left:"slide-in-from-left-transition",top:"slide-in-from-top-transition",bottom:"slide-in-from-bottom-transition"})[o.placement]),handleAfterLeave:D,bodyDirectives:f,handleMousedownResizeTrigger:C,handleMouseenterResizeTrigger:T,handleMouseleaveResizeTrigger:V,isDragging:p,isHoverOnResizeTrigger:y}},render(){const{$slots:o,mergedClsPrefix:l}=this;return this.displayDirective==="show"||this.displayed||this.show?Ie((b(),M("div",br,[(b(),L(vo,{disabled:!this.showMask||!this.trapFocus,active:this.show,autoFocus:this.autoFocus,onEsc:this.onEsc},{default:()=>(b(),L(Et,{name:this.transitionName,appear:this.isMounted,onAfterEnter:this.onAfterEnter,onAfterLeave:this.handleAfterLeave},{default:()=>Ie(bo("div",qe(this.$attrs,{role:"dialog",ref:"bodyRef","aria-modal":"true",class:[`${l}-drawer`,this.rtlEnabled&&`${l}-drawer--rtl`,`${l}-drawer--${this.placement}-placement`,this.isDragging&&`${l}-drawer--unselectable`,this.nativeScrollbar&&`${l}-drawer--native-scrollbar`]}),[this.resizable?(b(),M("div",{key:2,class:q([`${l}-drawer__resize-trigger`,(this.isDragging||this.isHoverOnResizeTrigger)&&`${l}-drawer__resize-trigger--hover`]),onMouseenter:this.handleMouseenterResizeTrigger,onMouseleave:this.handleMouseleaveResizeTrigger,onMousedown:this.handleMousedownResizeTrigger},null,42,fr)):null,this.nativeScrollbar?(b(),M("div",{key:3,class:q([`${l}-drawer-content-wrapper`,this.contentClass]),style:ce(this.contentStyle),role:"none"},[Z(()=>{var n;return(n=o.default)==null?void 0:n.call(o)})],6)):(b(),L(Lt,qe({key:4},this.scrollbarProps,{contentStyle:this.contentStyle,contentClass:[`${l}-drawer-content-wrapper`,this.contentClass],theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar}),ht(o),1040,["contentStyle","contentClass","theme","themeOverrides"]))]),this.bodyDirectives)},1032,["name","appear","onAfterEnter","onAfterLeave"]))},1032,["disabled","active","autoFocus","onEsc"]))])),[[Oe,this.displayDirective==="if"||this.displayed||this.show]]):null}});const{cubicBezierEaseIn:hr,cubicBezierEaseOut:pr}=et;function mr({duration:o="0.3s",leaveDuration:l="0.2s",name:n="slide-in-from-bottom"}={}){return[x(`&.${n}-transition-leave-active`,{transition:`transform ${l} ${hr}`}),x(`&.${n}-transition-enter-active`,{transition:`transform ${o} ${pr}`}),x(`&.${n}-transition-enter-to`,{transform:"translateY(0)"}),x(`&.${n}-transition-enter-from`,{transform:"translateY(100%)"}),x(`&.${n}-transition-leave-from`,{transform:"translateY(0)"}),x(`&.${n}-transition-leave-to`,{transform:"translateY(100%)"})]}const{cubicBezierEaseIn:gr,cubicBezierEaseOut:_r}=et;function yr({duration:o="0.3s",leaveDuration:l="0.2s",name:n="slide-in-from-left"}={}){return[x(`&.${n}-transition-leave-active`,{transition:`transform ${l} ${gr}`}),x(`&.${n}-transition-enter-active`,{transition:`transform ${o} ${_r}`}),x(`&.${n}-transition-enter-to`,{transform:"translateX(0)"}),x(`&.${n}-transition-enter-from`,{transform:"translateX(-100%)"}),x(`&.${n}-transition-leave-from`,{transform:"translateX(0)"}),x(`&.${n}-transition-leave-to`,{transform:"translateX(-100%)"})]}const{cubicBezierEaseIn:kr,cubicBezierEaseOut:xr}=et;function wr({duration:o="0.3s",leaveDuration:l="0.2s",name:n="slide-in-from-right"}={}){return[x(`&.${n}-transition-leave-active`,{transition:`transform ${l} ${kr}`}),x(`&.${n}-transition-enter-active`,{transition:`transform ${o} ${xr}`}),x(`&.${n}-transition-enter-to`,{transform:"translateX(0)"}),x(`&.${n}-transition-enter-from`,{transform:"translateX(100%)"}),x(`&.${n}-transition-leave-from`,{transform:"translateX(0)"}),x(`&.${n}-transition-leave-to`,{transform:"translateX(100%)"})]}const{cubicBezierEaseIn:Sr,cubicBezierEaseOut:Cr}=et;function zr({duration:o="0.3s",leaveDuration:l="0.2s",name:n="slide-in-from-top"}={}){return[x(`&.${n}-transition-leave-active`,{transition:`transform ${l} ${Sr}`}),x(`&.${n}-transition-enter-active`,{transition:`transform ${o} ${Cr}`}),x(`&.${n}-transition-enter-to`,{transform:"translateY(0)"}),x(`&.${n}-transition-enter-from`,{transform:"translateY(-100%)"}),x(`&.${n}-transition-leave-from`,{transform:"translateY(0)"}),x(`&.${n}-transition-leave-to`,{transform:"translateY(-100%)"})]}var Rr=x([U("drawer",`
 word-break: break-word;
 line-height: var(--n-line-height);
 position: absolute;
 pointer-events: all;
 box-shadow: var(--n-box-shadow);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background-color: var(--n-color);
 color: var(--n-text-color);
 box-sizing: border-box;
 `,[wr(),yr(),zr(),mr(),Y("unselectable",`
 user-select: none; 
 -webkit-user-select: none;
 `),Y("native-scrollbar",[U("drawer-content-wrapper",`
 overflow: auto;
 height: 100%;
 `)]),N("resize-trigger",`
 position: absolute;
 background-color: #0000;
 transition: background-color .3s var(--n-bezier);
 `,[Y("hover",`
 background-color: var(--n-resize-trigger-color-hover);
 `)]),U("drawer-content-wrapper",`
 box-sizing: border-box;
 `),U("drawer-content",`
 height: 100%;
 display: flex;
 flex-direction: column;
 `,[Y("native-scrollbar",[U("drawer-body-content-wrapper",`
 height: 100%;
 overflow: auto;
 `)]),U("drawer-body",`
 flex: 1 0 0;
 overflow: hidden;
 `),U("drawer-body-content-wrapper",`
 box-sizing: border-box;
 padding: var(--n-body-padding);
 `),U("drawer-header",`
 font-weight: var(--n-title-font-weight);
 line-height: 1;
 font-size: var(--n-title-font-size);
 color: var(--n-title-text-color);
 padding: var(--n-header-padding);
 transition: border .3s var(--n-bezier);
 border-bottom: 1px solid var(--n-divider-color);
 border-bottom: var(--n-header-border-bottom);
 display: flex;
 justify-content: space-between;
 align-items: center;
 `,[N("main",`
 flex: 1;
 `),N("close",`
 margin-left: 6px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)]),U("drawer-footer",`
 display: flex;
 justify-content: flex-end;
 border-top: var(--n-footer-border-top);
 transition: border .3s var(--n-bezier);
 padding: var(--n-footer-padding);
 `)]),Y("right-placement",`
 top: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-bottom-left-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 left: 0;
 transform: translateX(-1.5px);
 cursor: ew-resize;
 `)]),Y("left-placement",`
 top: 0;
 bottom: 0;
 left: 0;
 border-top-right-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 right: 0;
 transform: translateX(1.5px);
 cursor: ew-resize;
 `)]),Y("top-placement",`
 top: 0;
 left: 0;
 right: 0;
 border-bottom-left-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 100%;
 height: 3px;
 bottom: 0;
 left: 0;
 transform: translateY(1.5px);
 cursor: ns-resize;
 `)]),Y("bottom-placement",`
 left: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-top-right-radius: var(--n-border-radius);
 `,[N("resize-trigger",`
 width: 100%;
 height: 3px;
 top: 0;
 left: 0;
 transform: translateY(-1.5px);
 cursor: ns-resize;
 `)])]),x("body",[x(">",[U("drawer-container",`
 position: fixed;
 `)])]),U("drawer-container",`
 position: relative;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 pointer-events: none;
 `,[x("> *",`
 pointer-events: all;
 `)]),U("drawer-mask",`
 background-color: rgba(0, 0, 0, .3);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `,[Y("invisible",`
 background-color: rgba(0, 0, 0, 0)
 `),xo({enterDuration:"0.2s",leaveDuration:"0.2s",enterCubicBezier:"var(--n-bezier-in)",leaveCubicBezier:"var(--n-bezier-out)"})])]);const $r=["onClick"],Tr={...be.props,show:Boolean,width:[Number,String],height:[Number,String],placement:{type:String,default:"right"},maskClosable:{type:Boolean,default:!0},showMask:{type:[Boolean,String],default:!0},to:[String,Object],displayDirective:{type:String,default:"if"},nativeScrollbar:{type:Boolean,default:!0},zIndex:Number,onMaskClick:Function,scrollbarProps:Object,contentClass:String,contentStyle:[Object,String],trapFocus:{type:Boolean,default:!0},onEsc:Function,autoFocus:{type:Boolean,default:!0},closeOnEsc:{type:Boolean,default:!0},blockScroll:{type:Boolean,default:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,defaultWidth:{type:[Number,String],default:251},defaultHeight:{type:[Number,String],default:251},onUpdateWidth:[Function,Array],onUpdateHeight:[Function,Array],"onUpdate:width":[Function,Array],"onUpdate:height":[Function,Array],"onUpdate:show":[Function,Array],onUpdateShow:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,drawerStyle:[String,Object],drawerClass:String,target:null,onShow:Function,onHide:Function};var Br=fe({name:"Drawer",inheritAttrs:!1,props:Tr,setup(o){const{mergedClsPrefixRef:l,namespaceRef:n,inlineThemeDisabled:v}=Re(o),h=Co(),z=be("Drawer","-drawer",Rr,ur,o,l),g=$(o.defaultWidth),y=$(o.defaultHeight),p=Pe(ue(o,"width"),g),F=Pe(ue(o,"height"),y),I=X(()=>{const{placement:u}=o;return u==="top"||u==="bottom"?"":Rt(p.value)}),s=X(()=>{const{placement:u}=o;return u==="left"||u==="right"?"":Rt(F.value)}),m=u=>{const{onUpdateWidth:f,"onUpdate:width":D}=o;f&&G(f,u),D&&G(D,u),g.value=u},k=u=>{const{onUpdateHeight:f,"onUpdate:width":D}=o;f&&G(f,u),D&&G(D,u),y.value=u},C=X(()=>[{width:I.value,height:s.value},o.drawerStyle||""]);function T(u){const{onMaskClick:f,maskClosable:D}=o;D&&O(!1),f&&f(u)}function V(u){T(u)}const W=zo();function R(u){var f;(f=o.onEsc)==null||f.call(o),o.show&&o.closeOnEsc&&Ro(u)&&(W.value||O(!1))}function O(u){const{onHide:f,onUpdateShow:D,"onUpdate:show":_}=o;D&&G(D,u),_&&G(_,u),f&&!u&&G(f,u)}Fe(pt,{isMountedRef:h,mergedThemeRef:z,mergedClsPrefixRef:l,doUpdateShow:O,doUpdateHeight:k,doUpdateWidth:m});const B=X(()=>{const{common:{cubicBezierEaseInOut:u,cubicBezierEaseIn:f,cubicBezierEaseOut:D},self:{color:_,textColor:E,boxShadow:H,lineHeight:j,headerPadding:Q,footerPadding:le,borderRadius:ne,bodyPadding:ve,titleFontSize:he,titleTextColor:pe,titleFontWeight:me,headerBorderBottom:ie,footerBorderTop:ge,closeIconColor:_e,closeIconColorHover:ye,closeIconColorPressed:ke,closeColorHover:xe,closeColorPressed:Ue,closeIconSize:c,closeSize:Ee,closeBorderRadius:$e,resizableTriggerColorHover:we}}=z.value;return{"--n-line-height":j,"--n-color":_,"--n-border-radius":ne,"--n-text-color":E,"--n-box-shadow":H,"--n-bezier":u,"--n-bezier-out":D,"--n-bezier-in":f,"--n-header-padding":Q,"--n-body-padding":ve,"--n-footer-padding":le,"--n-title-text-color":pe,"--n-title-font-size":he,"--n-title-font-weight":me,"--n-header-border-bottom":ie,"--n-footer-border-top":ge,"--n-close-icon-color":_e,"--n-close-icon-color-hover":ye,"--n-close-icon-color-pressed":ke,"--n-close-size":Ee,"--n-close-color-hover":xe,"--n-close-color-pressed":Ue,"--n-close-icon-size":c,"--n-close-border-radius":$e,"--n-resize-trigger-color-hover":we}}),w=v?Ze("drawer",void 0,B,o):void 0;return{mergedClsPrefix:l,namespace:n,mergedBodyStyle:C,handleOutsideClick:V,handleMaskClick:T,handleEsc:R,mergedTheme:z,cssVars:v?void 0:B,themeClass:w==null?void 0:w.themeClass,onRender:w==null?void 0:w.onRender,isMounted:h}},render(){const{mergedClsPrefix:o}=this;return b(),L(So,{to:this.to,show:this.show},{default:()=>{var l;return(l=this.onRender)==null||l.call(this),Ie((b(),M("div",{class:q([`${o}-drawer-container`,this.namespace,this.themeClass]),style:ce(this.cssVars),role:"none"},[this.showMask?(b(),L(Et,{key:0,name:"fade-in-transition",appear:this.isMounted},{default:()=>this.show?(b(),M("div",{key:1,"aria-hidden":!0,class:q([`${o}-drawer-mask`,this.showMask==="transparent"&&`${o}-drawer-mask--invisible`]),onClick:this.handleMaskClick},null,10,$r)):null},1032,["appear"])):Z(()=>null),(b(),L(vr,qe(this.$attrs,{class:[this.drawerClass,this.$attrs.class],style:[this.mergedBodyStyle,this.$attrs.style],blockScroll:this.blockScroll,contentStyle:this.contentStyle,contentClass:this.contentClass,placement:this.placement,scrollbarProps:this.scrollbarProps,show:this.show,displayDirective:this.displayDirective,nativeScrollbar:this.nativeScrollbar,onAfterEnter:this.onAfterEnter,onAfterLeave:this.onAfterLeave,trapFocus:this.trapFocus,autoFocus:this.autoFocus,resizable:this.resizable,maxHeight:this.maxHeight,minHeight:this.minHeight,maxWidth:this.maxWidth,minWidth:this.minWidth,showMask:this.showMask,onEsc:this.handleEsc,onClickoutside:this.handleOutsideClick}),ht(this.$slots),1040,["class","style","blockScroll","contentStyle","contentClass","placement","scrollbarProps","show","displayDirective","nativeScrollbar","onAfterEnter","onAfterLeave","trapFocus","autoFocus","resizable","maxHeight","minHeight","maxWidth","minWidth","showMask","onEsc","onClickoutside"]))],6)),[[wo,{zIndex:this.zIndex,enabled:this.show}]])}},1032,["to","show"])}});const Fr={title:String,headerClass:String,headerStyle:[Object,String],footerClass:String,footerStyle:[Object,String],bodyClass:String,bodyStyle:[Object,String],bodyContentClass:String,bodyContentStyle:[Object,String],nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,closable:Boolean};var Ir=fe({name:"DrawerContent",props:Fr,slots:Object,setup(){const o=Xe(pt,null);o||$o("drawer-content","`n-drawer-content` must be placed inside `n-drawer`.");const{doUpdateShow:l}=o;function n(){l(!1)}return{handleCloseClick:n,mergedTheme:o.mergedThemeRef,mergedClsPrefix:o.mergedClsPrefixRef}},render(){const{title:o,mergedClsPrefix:l,nativeScrollbar:n,mergedTheme:v,bodyClass:h,bodyStyle:z,bodyContentClass:g,bodyContentStyle:y,headerClass:p,headerStyle:F,footerClass:I,footerStyle:s,scrollbarProps:m,closable:k,$slots:C}=this;return b(),M("div",{role:"none",class:q([`${l}-drawer-content`,n&&`${l}-drawer-content--native-scrollbar`])},[C.header||o||k?(b(),M("div",{key:0,class:q([`${l}-drawer-header`,p]),style:ce(F),role:"none"},[P("div",{class:q(`${l}-drawer-header__main`),role:"heading","aria-level":"1"},[C.header!==void 0?(b(),M(re,{key:0},[Z(()=>C.header())],64)):(b(),M(re,{key:1},[Z(()=>o)],64))],2),Z(()=>k&&(b(),L(To,{onClick:this.handleCloseClick,clsPrefix:l,class:q(`${l}-drawer-header__close`),absolute:!0},null,8,["onClick","clsPrefix","class"])))],6)):Z(()=>null),n?(b(),M("div",{key:2,class:q([`${l}-drawer-body`,h]),style:ce(z),role:"none"},[P("div",{class:q([`${l}-drawer-body-content-wrapper`,g]),style:ce(y),role:"none"},[Z(()=>{var T;return(T=C.default)==null?void 0:T.call(C)})],6)],6)):(b(),L(Lt,qe({key:3,themeOverrides:v.peerOverrides.Scrollbar,theme:v.peers.Scrollbar},m,{class:`${l}-drawer-body`,contentClass:[`${l}-drawer-body-content-wrapper`,g],contentStyle:y}),ht(C),1040,["themeOverrides","theme","class","contentClass","contentStyle"])),C.footer?(b(),M("div",{key:4,class:q([`${l}-drawer-footer`,I]),style:ce(s),role:"none"},[Z(()=>C.footer())],6)):Z(()=>null)],2)}});const ct=$({});let Dt=!1,Ae=null;function Dr(){async function o(n=!1){if(!(Dt&&!n))return Ae&&!n||(Ae=(async()=>{const v={};try{const h=await Promise.allSettled([de.listAvailableCerts({purpose:"sign",protocol:"http"}),de.listAvailableCerts({purpose:"sign",protocol:"gmssl"}),de.listAvailableCerts({purpose:"enc",protocol:"gmssl"})]);for(const z of h)if(z.status==="fulfilled")for(const g of z.value.data.items||[])v[g.cert_id]||(v[g.cert_id]=g.subject_cn||g.cert_id);ct.value=v,Dt=!0}catch{}finally{Ae=null}})()),Ae}function l(n){return n?ct.value[n]||n:""}return{certMap:ct,load:o,certName:l}}const Pr={class:"drawer-header"},Ur={class:"mono"},Er={class:"mono"},Lr={key:1},Ar={class:"mono"},Mr={class:"form-hint"},Or={class:"form-hint"},Hr={class:"form-hint"},Nr={class:"mono",style:{"word-break":"break-all"}},Vr={class:"form-hint"},jr={class:"form-hint"},Wr={class:"form-hint"},Gr={class:"form-hint"},Yr={key:1,class:"ciphers-group"},qr={class:"ciphers-group__section"},Kr={class:"ciphers-group__section"},Xr={class:"ciphers-group__section"},Jr={class:"ciphers-group__preview"},Qr={class:"mono ciphers-group__preview-text"},Zr=fe({__name:"ProxyDetailDrawer",props:{show:{type:Boolean},serviceId:{}},emits:["update:show","saved"],setup(o,{emit:l}){const n=o,v=l,h=Po(),{load:z,certName:g}=Dr(),y=Bo(),p=t=>y.hasPermission(t),F=$(!1),I=$(!1),s=$(null),m=$(!1),k=$("basic"),C=$("logs"),T=$(null),V=$(null),W=$(null),R=$("intl"),O=$(!1),B=$([]),w=$("default"),u=$([]),f=$(""),D=$(!1),_=$(!1),E=$(!0),H=$(!1),j=$([]),Q=$(!1),le=X({get:()=>c.listen_ssl!==!1,set:t=>{c.listen_ssl=t}}),ne=X({get:()=>c.backend_scheme==="https",set:t=>{c.backend_scheme=t?"https":"http"}}),ve=X(()=>Oo.filter(t=>t.value===Bt?c.protocol==="gmssl":!0)),he=it.filter(t=>t.group==="tls13"),pe=it.filter(t=>t.group==="tls12"),me=it.filter(t=>t.group==="gm"),ie=$([]),ge=$([]),_e=$([]),ye=$(!1),ke=$(!1),xe=$(!1);function Ue(t="client"){const e="http";return{service_name:"",mode:t,protocol:e,listen_ip:"0.0.0.0",listen_port:8080,backend_ip:"127.0.0.1",backend_port:8080,backend_scheme:t==="client"?"https":"http",ssl_server_name:"",listen_ssl:t==="server"?!0:void 0,cert_id:"",enc_cert_id:"",intl_cert_id:"",ca_cert_id:"",ca_domain:"",client_cert_id:"",client_enc_cert_id:"",intl_client_cert_id:"",client_key_ref:"",ssl_protocols:dt(e),ssl_ciphers:Ce(),ssl_conf_command:Ao(e),ssl_ciphers_mode:"default",ssl_session_cache_enabled:t==="server"?!1:void 0,ssl_verify_client:t==="server"?!0:void 0,ssl_session_tickets:t==="server"?!1:void 0}}const c=Fo(Ue()),Ee={service_name:{required:!0,message:"请填写配置名称",trigger:"blur"},listen_ip:{required:!0,message:"请选择本地监听地址",trigger:["blur","change"]},listen_port:{required:!0,type:"number",message:"请填写本地端口",trigger:["blur","change"]},backend_ip:{required:!0,message:"请填写后台服务地址",trigger:"blur"},backend_port:{required:!0,type:"number",message:"请填写后台服务端口",trigger:["blur","change"]}},$e=X(()=>O.value),we=X(()=>O.value),tt=X(()=>c.mode==="server"&&le.value),ot=X(()=>c.mode==="server"&&le.value?"开启 SSL 通信时必须配置证书":O.value?"本代理加载证书":"本代理不加载证书");Ge(le,t=>{c.mode==="server"&&t&&(O.value=!0)}),Ge(w,t=>{t==="input"&&(f.value.trim()||(f.value=Ft))});async function mt(){var t;if(!(j.value.length>0)){Q.value=!0;try{const i=(await de.listCADomains()).data;let J=[];Array.isArray(i==null?void 0:i.domains)?J=i.domains:Array.isArray((t=i==null?void 0:i.data)==null?void 0:t.domains)?J=i.data.domains:Array.isArray(i)&&(J=i),J.includes("safe")||J.unshift("safe"),j.value=J.map(ae=>({label:ae,value:ae}))}catch(e){h.error(e.message||"加载 CA 信任域失败"),j.value=[{label:"safe",value:"safe"}]}finally{Q.value=!1}}}function gt(){c.intl_cert_id="",c.cert_id="",c.enc_cert_id=""}function Ht(t){t==="intl"?(c.cert_id="",c.enc_cert_id=""):t==="gmssl"&&(c.intl_cert_id="");const e=It(t),i=dt(e);c.ssl_protocols=i,B.value=st(i),w.value==="default"&&(c.ssl_ciphers=Ce(c.mode)),xt(t)}function rt(t){return t.map(e=>({label:`${e.subject_cn||e.cert_id}（${e.algorithm}）`,value:e.cert_id}))}async function _t(){if(!(ie.value.length>0)){ye.value=!0;try{const t=await de.listAvailableCerts({purpose:"sign",protocol:"http"});ie.value=rt(t.data.items||[])}catch(t){h.error(t.message||"加载国际签名证书失败")}finally{ye.value=!1}}}async function yt(){if(!(ge.value.length>0)){ke.value=!0;try{const t=await de.listAvailableCerts({purpose:"sign",protocol:"gmssl"});ge.value=rt(t.data.items||[])}catch(t){h.error(t.message||"加载国密签名证书失败")}finally{ke.value=!1}}}async function kt(){if(!(_e.value.length>0)){xe.value=!0;try{const t=await de.listAvailableCerts({purpose:"enc",protocol:"gmssl"});_e.value=rt(t.data.items||[])}catch(t){h.error(t.message||"加载国密加密证书失败")}finally{xe.value=!1}}}function xt(t){t==="intl"?_t():t==="gmssl"?(yt(),kt()):(_t(),yt(),kt())}function lt(t){return t.listen_ssl!==!1}function Nt(t){var i;const e=ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id));return((i=Tt.find(J=>J.value===e))==null?void 0:i.label)||e}function Le(t){return!!(t.intl_cert_id||t.intl_client_cert_id||t.cert_id||t.client_cert_id||t.enc_cert_id||t.client_enc_cert_id)}function Vt(t){return t.mode==="server"?lt(t)&&Le(t):Le(t)}function jt(t){const e=ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id));return e==="intl"||e==="intl_gmssl"}function wt(t){const e=ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id));return e==="gmssl"||e==="intl_gmssl"}function Wt(t){return wt(t)}function Gt(t){return ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id))==="intl"?t.mode==="client"?t.client_cert_id||"":t.cert_id||"":t.mode==="client"?t.intl_client_cert_id||"":t.intl_cert_id||""}function Yt(t){return ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id))==="intl"?"":t.mode==="client"?t.client_cert_id||"":t.cert_id||""}function qt(t){return ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id))==="intl"?"":t.mode==="client"?t.client_enc_cert_id||"":t.enc_cert_id||""}function Kt(t){const e=st(t);return e.length===0?"—":e.map(i=>i===Bt?"GMTLSv1.1":i).join(" ")}async function St(t){F.value=!0;try{const e=await de.get(t);s.value=e.data}catch(e){h.error(e.message||"加载失败"),v("update:show",!1)}finally{F.value=!1}}function He(){if(!s.value)return;const t=s.value;Object.assign(c,{service_name:t.service_name,mode:t.mode,protocol:t.protocol,listen_ip:t.listen_ip,listen_port:t.listen_port,backend_ip:t.backend_ip,backend_port:t.backend_port,backend_scheme:t.backend_scheme,ssl_server_name:t.ssl_server_name,listen_ssl:t.listen_ssl!==!1,ca_cert_id:t.ca_cert_id||"",ca_domain:t.ca_domain||"",client_key_ref:t.client_key_ref,ssl_protocols:t.ssl_protocols||dt(t.protocol),ssl_ciphers:t.ssl_ciphers||Ce(t.mode),ssl_conf_command:t.ssl_conf_command||"",ssl_ciphers_mode:t.ssl_ciphers_mode||"default",ssl_session_cache_enabled:t.ssl_session_cache_enabled??(t.mode==="server"?!1:void 0),ssl_verify_client:t.ssl_verify_client??(t.mode==="server"?!0:void 0),ssl_session_tickets:t.ssl_session_tickets??(t.mode==="server"?!1:void 0)}),B.value=st(c.ssl_protocols),w.value=t.ssl_ciphers_mode||"default",w.value==="custom"?(u.value=(t.ssl_ciphers||"").split(":").map(Ne=>Ne.trim()).filter(Boolean),f.value=""):w.value==="input"?(u.value=[],f.value=(t.ssl_ciphers||"").trim()||Ft):(u.value=[],f.value=""),D.value=!!t.ssl_session_cache_enabled,_.value=!!t.ssl_session_cache_enabled,E.value=t.ssl_verify_client!==!1,H.value=!!t.ssl_session_tickets,gt(),R.value=ze(t.protocol,!!(t.intl_cert_id||t.intl_client_cert_id));const e=t.mode==="client",i=e?t.client_cert_id||"":t.cert_id||"",J=e?t.client_enc_cert_id||"":t.enc_cert_id||"",ae=e?t.intl_client_cert_id||"":t.intl_cert_id||"";R.value==="intl"?c.intl_cert_id=i:R.value==="gmssl"?(c.cert_id=i,c.enc_cert_id=J):(c.intl_cert_id=ae,c.cert_id=i,c.enc_cert_id=J),!e&&t.listen_ssl!==!1?O.value=!0:O.value=Le(t)}Ge(()=>[n.show,n.serviceId],async([t,e])=>{var i,J,ae;if(!t){m.value=!1,k.value="basic",s.value=null,R.value="intl",O.value=!1,B.value=[],w.value="default",u.value=[],f.value="",gt();return}m.value=!1,k.value="basic",p("ssl_proxy:proxy:config")?C.value="kv":C.value="logs",e&&(await Promise.all([St(e),z(),mt()]),He()),await Ct(),(i=T.value)==null||i.restoreValidation(),(J=V.value)==null||J.restoreValidation(),(ae=W.value)==null||ae.restoreValidation()},{immediate:!0});function Xt(t){v("update:show",t)}function Jt(){s.value&&(He(),m.value=!0,$e.value&&xt(R.value),mt(),Ct(()=>{var t,e,i;(t=T.value)==null||t.restoreValidation(),(e=V.value)==null||e.restoreValidation(),(i=W.value)==null||i.restoreValidation()}))}function Qt(){m.value=!1,s.value&&He()}function Zt(){if(!O.value)return null;if(R.value==="intl"){if(!c.intl_cert_id)return"请选择国际签名证书"}else if(R.value==="gmssl"){if(!c.cert_id)return"请选择国密签名证书";if(!c.enc_cert_id)return"请选择国密加密证书"}else{if(!c.intl_cert_id)return"请选择国际签名证书";if(!c.cert_id)return"请选择国密签名证书";if(!c.enc_cert_id)return"请选择国密加密证书"}return null}function eo(){var i;const t=It(R.value),e={...c,protocol:t,service_name:((i=s.value)==null?void 0:i.service_name)||c.service_name};if(e.mode==="client"&&(e.listen_ssl=void 0),e.cert_id="",e.enc_cert_id="",e.intl_cert_id="",e.client_cert_id="",e.client_enc_cert_id="",e.intl_client_cert_id="",O.value){const J=e.mode==="client";R.value==="intl"?J?e.client_cert_id=c.intl_cert_id||"":e.cert_id=c.intl_cert_id||"":R.value==="gmssl"?J?(e.client_cert_id=c.cert_id||"",e.client_enc_cert_id=c.enc_cert_id||""):(e.cert_id=c.cert_id||"",e.enc_cert_id=c.enc_cert_id||""):J?(e.intl_client_cert_id=c.intl_cert_id||"",e.client_cert_id=c.cert_id||"",e.client_enc_cert_id=c.enc_cert_id||""):(e.intl_cert_id=c.intl_cert_id||"",e.cert_id=c.cert_id||"",e.enc_cert_id=c.enc_cert_id||"")}return e.ssl_protocols=Mo(B.value,t),w.value==="default"?e.ssl_ciphers=Ce(e.mode):w.value==="custom"?e.ssl_ciphers=u.value.join(":"):e.ssl_ciphers=f.value.trim(),e.ssl_ciphers_mode=w.value,e.ssl_ciphers_mode!=="default"&&!(e.ssl_ciphers||"").trim()&&(e.ssl_ciphers_mode="default",e.ssl_ciphers=Ce(e.mode)),e.mode==="client"?(e.ssl_session_cache_enabled=D.value,e.ssl_verify_client=void 0,e.ssl_session_tickets=void 0):(e.ssl_session_cache_enabled=_.value,e.ssl_verify_client=E.value,e.ssl_session_tickets=H.value),e}async function to(){var J,ae,Ne;try{k.value==="basic"?await((J=T.value)==null?void 0:J.validate()):k.value==="backend"?await((ae=V.value)==null?void 0:ae.validate()):await((Ne=W.value)==null?void 0:Ne.validate())}catch{return}if(!n.serviceId)return;const t=Zt();if(t){h.error(t);return}const e=!!s.value&&s.value.status==="RUNNING",i=eo();I.value=!0;try{await de.update(n.serviceId,i),m.value=!1,await Promise.all([St(n.serviceId),z(!0)]),He(),v("saved"),e?h.warning("配置已修改，请重启服务使其生效"):h.success("保存成功")}catch(oo){h.error(oo.message||"保存失败")}finally{I.value=!1}}return(t,e)=>(b(),L(r(Br),{show:o.show,width:760,placement:"right","mask-closable":!1,"onUpdate:show":Xt},{default:a(()=>[d(r(Ir),{closable:""},{header:a(()=>[P("div",Pr,[d(r(oe),{align:"center"},{default:a(()=>[P("span",null,A(m.value?"编辑配置":"配置详情"),1),s.value?(b(),L(r(se),{key:0,type:r(Lo)[s.value.status]||"default",size:"small"},{default:a(()=>[S(A(s.value.status),1)]),_:1},8,["type"])):te("",!0)]),_:1}),d(r(oe),{class:"drawer-header__actions"},{default:a(()=>[m.value?(b(),M(re,{key:1},[d(r(Te),{size:"small",onClick:Qt},{default:a(()=>[...e[27]||(e[27]=[S("取消",-1)])]),_:1}),d(r(Te),{size:"small",type:"primary",loading:I.value,onClick:to},{default:a(()=>[...e[28]||(e[28]=[S(" 保存 ",-1)])]),_:1},8,["loading"])],64)):(b(),M(re,{key:0},[p("ssl_proxy:proxy:update")?(b(),L(r(Te),{key:0,size:"small",type:"primary",onClick:Jt},{default:a(()=>[...e[26]||(e[26]=[S(" 编辑 ",-1)])]),_:1})):te("",!0)],64))]),_:1})])]),default:a(()=>[d(r(Do),{show:F.value},{default:a(()=>[d(r(oe),{style:{"margin-bottom":"16px"}},{default:a(()=>[d(r(Te),{type:k.value==="basic"?"primary":"default",size:"small",onClick:e[0]||(e[0]=i=>k.value="basic")},{default:a(()=>[...e[29]||(e[29]=[S(" 基础配置 ",-1)])]),_:1},8,["type"]),d(r(Te),{type:k.value==="backend"?"primary":"default",size:"small",onClick:e[1]||(e[1]=i=>k.value="backend")},{default:a(()=>[...e[30]||(e[30]=[S(" 后端代理配置 ",-1)])]),_:1},8,["type"]),d(r(Te),{type:k.value==="ssl"?"primary":"default",size:"small",onClick:e[2]||(e[2]=i=>k.value="ssl")},{default:a(()=>[...e[31]||(e[31]=[S(" SSL配置 ",-1)])]),_:1},8,["type"])]),_:1}),Ie(P("div",null,[!m.value&&s.value?(b(),L(r(nt),{key:0,bordered:"",column:1,"label-placement":"left",size:"small"},{default:a(()=>[d(r(K),{label:"配置名称"},{default:a(()=>[S(A(s.value.service_name),1)]),_:1}),d(r(K),{label:"本地监听地址"},{default:a(()=>[P("span",Ur,A(s.value.listen_ip),1)]),_:1}),d(r(K),{label:"本地端口"},{default:a(()=>[S(A(s.value.listen_port),1)]),_:1}),s.value.mode==="server"?(b(),L(r(K),{key:0,label:"开启SSL通信"},{default:a(()=>[d(r(se),{type:lt(s.value)?"success":"default",size:"small"},{default:a(()=>[S(A(lt(s.value)?"是":"否"),1)]),_:1},8,["type"])]),_:1})):te("",!0),d(r(K),{label:"后台服务地址"},{default:a(()=>[P("span",Er,A(s.value.backend_ip),1)]),_:1}),d(r(K),{label:"后台服务端口"},{default:a(()=>[S(A(s.value.backend_port),1)]),_:1}),d(r(K),{label:"证书类型"},{default:a(()=>[S(A(Nt(s.value)),1)]),_:1}),d(r(K),{label:"配置证书"},{default:a(()=>[d(r(se),{type:Le(s.value)?"success":"default",size:"small"},{default:a(()=>[S(A(Le(s.value)?"是":"否"),1)]),_:1},8,["type"])]),_:1}),Vt(s.value)?(b(),M(re,{key:1},[jt(s.value)?(b(),L(r(K),{key:0,label:"国际签名证书"},{default:a(()=>[S(A(r(g)(Gt(s.value))||"—"),1)]),_:1})):te("",!0),wt(s.value)?(b(),L(r(K),{key:1,label:"国密签名证书"},{default:a(()=>[S(A(r(g)(Yt(s.value))||"—"),1)]),_:1})):te("",!0),Wt(s.value)?(b(),L(r(K),{key:2,label:"国密加密证书"},{default:a(()=>[S(A(r(g)(qt(s.value))||"—"),1)]),_:1})):te("",!0)],64)):te("",!0),d(r(K),{label:"信任域(验证服务端证书)"},{default:a(()=>[s.value.ca_domain?(b(),L(r(se),{key:0,size:"small",type:"info"},{default:a(()=>[S(A(s.value.ca_domain),1)]),_:1})):(b(),M("span",Lr,"—"))]),_:1}),d(r(K),{label:"代理模式"},{default:a(()=>[S(A(s.value.mode==="client"?"客户端":"服务端"),1)]),_:1}),d(r(K),{label:"service_id"},{default:a(()=>[P("span",Ar,A(s.value.service_id),1)]),_:1}),d(r(K),{label:"PID"},{default:a(()=>[S(A(s.value.pid>0?s.value.pid:"—"),1)]),_:1}),d(r(K),{label:"创建时间"},{default:a(()=>[S(A(s.value.created_at),1)]),_:1}),d(r(K),{label:"更新时间"},{default:a(()=>[S(A(s.value.updated_at),1)]),_:1})]),_:1})):(b(),L(r(at),{key:1,ref_key:"formRef",ref:T,model:c,rules:Ee,"label-placement":"left","label-width":"160"},{default:a(()=>[d(r(ee),{label:"配置名称",path:"service_name"},{default:a(()=>[d(r(je),{value:c.service_name,"onUpdate:value":e[3]||(e[3]=i=>c.service_name=i),placeholder:"请填写配置名称",maxlength:"64",disabled:""},null,8,["value"]),e[32]||(e[32]=P("span",{class:"form-hint-block"},"配置名称创建后不可修改",-1))]),_:1}),d(r(ee),{label:"本地监听地址",path:"listen_ip"},{default:a(()=>[d(r(Be),{value:c.listen_ip,"onUpdate:value":e[4]||(e[4]=i=>c.listen_ip=i),options:r(Uo),placeholder:"请选择",filterable:"",tag:""},null,8,["value","options"])]),_:1}),d(r(ee),{label:"本地端口",path:"listen_port"},{default:a(()=>[d(r(zt),{value:c.listen_port,"onUpdate:value":e[5]||(e[5]=i=>c.listen_port=i),min:1,max:65535,placeholder:"请填写本地端口",style:{width:"100%"}},null,8,["value"])]),_:1}),c.mode==="server"?(b(),L(r(ee),{key:0,label:"开启SSL通信"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:le.value,"onUpdate:value":e[6]||(e[6]=i=>le.value=i)},{checked:a(()=>[...e[33]||(e[33]=[S("是",-1)])]),unchecked:a(()=>[...e[34]||(e[34]=[S("否",-1)])]),_:1},8,["value"]),P("span",Mr,A(le.value?"服务端监听使用 SSL 加密，必须配置证书":"服务端监听使用明文 HTTP"),1)]),_:1})]),_:1})):te("",!0),d(r(ee),{label:"后台服务地址",path:"backend_ip"},{default:a(()=>[d(r(je),{value:c.backend_ip,"onUpdate:value":e[7]||(e[7]=i=>c.backend_ip=i),placeholder:"例如 127.0.0.1 或 0.0.0.0"},null,8,["value"])]),_:1}),d(r(ee),{label:"后台服务端口",path:"backend_port"},{default:a(()=>[d(r(zt),{value:c.backend_port,"onUpdate:value":e[8]||(e[8]=i=>c.backend_port=i),min:1,max:65535,placeholder:"请填写后台服务端口",style:{width:"100%"}},null,8,["value"])]),_:1}),d(r(ee),{label:"证书类型"},{default:a(()=>[d(r(Be),{value:R.value,"onUpdate:value":[e[9]||(e[9]=i=>R.value=i),Ht],options:r(Tt),placeholder:"请选择证书类型"},null,8,["value","options"])]),_:1}),d(r(ee),{label:"配置证书"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:O.value,"onUpdate:value":e[10]||(e[10]=i=>O.value=i),disabled:tt.value},{checked:a(()=>[...e[35]||(e[35]=[S("是",-1)])]),unchecked:a(()=>[...e[36]||(e[36]=[S("否",-1)])]),_:1},8,["value","disabled"]),P("span",Or,A(ot.value),1)]),_:1})]),_:1}),$e.value?(b(),M(re,{key:1},[R.value==="intl"||R.value==="intl_gmssl"?(b(),L(r(ee),{key:0,label:"国际签名证书",required:we.value},{default:a(()=>[d(r(Be),{value:c.intl_cert_id,"onUpdate:value":e[11]||(e[11]=i=>c.intl_cert_id=i),options:ie.value,loading:ye.value,placeholder:"请选择国际签名证书",clearable:"",filterable:""},null,8,["value","options","loading"])]),_:1},8,["required"])):te("",!0),R.value==="gmssl"||R.value==="intl_gmssl"?(b(),L(r(ee),{key:1,label:"国密签名证书",required:we.value},{default:a(()=>[d(r(Be),{value:c.cert_id,"onUpdate:value":e[12]||(e[12]=i=>c.cert_id=i),options:ge.value,loading:ke.value,placeholder:"请选择国密签名证书",clearable:"",filterable:""},null,8,["value","options","loading"])]),_:1},8,["required"])):te("",!0),R.value==="gmssl"||R.value==="intl_gmssl"?(b(),L(r(ee),{key:2,label:"国密加密证书",required:we.value},{default:a(()=>[d(r(Be),{value:c.enc_cert_id,"onUpdate:value":e[13]||(e[13]=i=>c.enc_cert_id=i),options:_e.value,loading:xe.value,placeholder:"请选择国密加密证书",clearable:"",filterable:""},null,8,["value","options","loading"])]),_:1},8,["required"])):te("",!0)],64)):te("",!0),d(r(ee),{label:"信任域(验证服务端证书)"},{default:a(()=>[d(r(Be),{value:c.ca_domain,"onUpdate:value":e[14]||(e[14]=i=>c.ca_domain=i),options:j.value,loading:Q.value,placeholder:"请选择 CA 信任域，留空表示不校验",clearable:"",filterable:""},null,8,["value","options","loading"])]),_:1})]),_:1},8,["model"]))],512),[[Oe,k.value==="basic"]]),Ie(P("div",null,[!m.value&&s.value?(b(),L(r(nt),{key:0,bordered:"",column:1,"label-placement":"left",size:"small"},{default:a(()=>[d(r(K),{label:"开启 SSL 通信"},{default:a(()=>[d(r(se),{type:s.value.backend_scheme==="https"?"success":"default",size:"small"},{default:a(()=>[S(A(s.value.backend_scheme==="https"?"是":"否"),1)]),_:1},8,["type"])]),_:1})]),_:1})):(b(),L(r(at),{key:1,ref_key:"backendFormRef",ref:V,model:c,"label-placement":"left","label-width":"140"},{default:a(()=>[d(r(ee),{label:"开启 SSL 通信"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:ne.value,"onUpdate:value":e[15]||(e[15]=i=>ne.value=i)},{checked:a(()=>[...e[37]||(e[37]=[S("是",-1)])]),unchecked:a(()=>[...e[38]||(e[38]=[S("否",-1)])]),_:1},8,["value"]),P("span",Hr,A(ne.value?"使用 HTTPS 访问后台服务":"使用 HTTP 访问后台服务"),1)]),_:1})]),_:1})]),_:1},8,["model"]))],512),[[Oe,k.value==="backend"]]),Ie(P("div",null,[!m.value&&s.value?(b(),L(r(nt),{key:0,bordered:"",column:1,"label-placement":"left",size:"small"},{default:a(()=>[d(r(K),{label:"SSL协议"},{default:a(()=>[S(A(Kt(s.value.ssl_protocols)),1)]),_:1}),s.value.mode==="client"?(b(),L(r(K),{key:0,label:"SSL会话缓存"},{default:a(()=>[d(r(se),{type:s.value.ssl_session_cache_enabled?"success":"default",size:"small"},{default:a(()=>[S(A(s.value.ssl_session_cache_enabled?"开启":"关闭"),1)]),_:1},8,["type"])]),_:1})):te("",!0),s.value.mode==="server"?(b(),M(re,{key:1},[d(r(K),{label:"SSL会话缓存"},{default:a(()=>[d(r(se),{type:s.value.ssl_session_cache_enabled?"success":"default",size:"small"},{default:a(()=>[S(A(s.value.ssl_session_cache_enabled?"开启":"关闭"),1)]),_:1},8,["type"])]),_:1}),d(r(K),{label:"SSL双向认证"},{default:a(()=>[d(r(se),{type:s.value.ssl_verify_client?"success":"default",size:"small"},{default:a(()=>[S(A(s.value.ssl_verify_client?"开启":"关闭"),1)]),_:1},8,["type"])]),_:1}),d(r(K),{label:"SSL会话票据"},{default:a(()=>[d(r(se),{type:s.value.ssl_session_tickets?"success":"default",size:"small"},{default:a(()=>[S(A(s.value.ssl_session_tickets?"开启":"关闭"),1)]),_:1},8,["type"])]),_:1})],64)):te("",!0),d(r(K),{label:"SSL加密套件"},{default:a(()=>[P("div",Nr,A(s.value.ssl_ciphers_mode==="default"||!s.value.ssl_ciphers?r(Ce)(s.value.mode):s.value.ssl_ciphers),1)]),_:1})]),_:1})):(b(),L(r(at),{key:1,ref_key:"sslFormRef",ref:W,model:c,"label-placement":"left","label-width":"140"},{default:a(()=>[d(r($t),{"title-placement":"left",style:{margin:"0 0 16px"}},{default:a(()=>[...e[39]||(e[39]=[S(" SSL协议配置 ",-1)])]),_:1}),d(r(ee),{label:"SSL协议"},{default:a(()=>[d(r(We),{value:B.value,"onUpdate:value":e[16]||(e[16]=i=>B.value=i)},{default:a(()=>[d(r(oe),null,{default:a(()=>[(b(!0),M(re,null,Ve(ve.value,i=>(b(),L(r(Me),{key:i.value,value:i.value,label:i.label},null,8,["value","label"]))),128))]),_:1})]),_:1},8,["value"])]),_:1}),c.mode==="client"?(b(),L(r(ee),{key:0,label:"SSL会话缓存"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:D.value,"onUpdate:value":e[17]||(e[17]=i=>D.value=i)},{checked:a(()=>[...e[40]||(e[40]=[S("开启",-1)])]),unchecked:a(()=>[...e[41]||(e[41]=[S("关闭",-1)])]),_:1},8,["value"]),P("span",Vr,A(D.value?"启用共享 SSL 会话缓存（shared:SSL:10m）":"不启用会话缓存"),1)]),_:1})]),_:1})):te("",!0),c.mode==="server"?(b(),M(re,{key:1},[d(r(ee),{label:"SSL会话缓存"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:_.value,"onUpdate:value":e[18]||(e[18]=i=>_.value=i)},{checked:a(()=>[...e[42]||(e[42]=[S("开启",-1)])]),unchecked:a(()=>[...e[43]||(e[43]=[S("关闭",-1)])]),_:1},8,["value"]),P("span",jr,A(_.value?"启用共享 SSL 会话缓存（shared:SSL:10m）":"不启用会话缓存"),1)]),_:1})]),_:1}),d(r(ee),{label:"SSL双向认证"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:E.value,"onUpdate:value":e[19]||(e[19]=i=>E.value=i)},{checked:a(()=>[...e[44]||(e[44]=[S("开启",-1)])]),unchecked:a(()=>[...e[45]||(e[45]=[S("关闭",-1)])]),_:1},8,["value"]),P("span",Wr,A(E.value?"要求客户端提供证书（需配置 CA 信任域）":"不校验客户端证书"),1)]),_:1})]),_:1}),d(r(ee),{label:"SSL会话票据"},{default:a(()=>[d(r(oe),{align:"center"},{default:a(()=>[d(r(Se),{value:H.value,"onUpdate:value":e[20]||(e[20]=i=>H.value=i)},{checked:a(()=>[...e[46]||(e[46]=[S("开启",-1)])]),unchecked:a(()=>[...e[47]||(e[47]=[S("关闭",-1)])]),_:1},8,["value"]),P("span",Gr,A(H.value?"允许会话票据":"禁止会话票据"),1)]),_:1})]),_:1})],64)):te("",!0),d(r($t),{"title-placement":"left",style:{margin:"16px 0"}},{default:a(()=>[...e[48]||(e[48]=[S(" SSL加密套件 ",-1)])]),_:1}),d(r(ee),{label:"SSL加密套件"},{default:a(()=>[d(r(oe),{vertical:"",style:{width:"100%"}},{default:a(()=>[d(r(dr),{value:w.value,"onUpdate:value":e[21]||(e[21]=i=>w.value=i)},{default:a(()=>[d(r(oe),null,{default:a(()=>[d(r(Ye),{value:"default"},{default:a(()=>[...e[49]||(e[49]=[S("默认套件",-1)])]),_:1}),d(r(Ye),{value:"custom"},{default:a(()=>[...e[50]||(e[50]=[S("自定义套件",-1)])]),_:1}),d(r(Ye),{value:"input"},{default:a(()=>[...e[51]||(e[51]=[S("输入套件",-1)])]),_:1})]),_:1})]),_:1},8,["value"]),w.value==="default"?(b(),L(r(je),{key:0,value:r(Ce)(c.mode),type:"textarea",autosize:{minRows:2,maxRows:4},readonly:""},null,8,["value"])):w.value==="custom"?(b(),M("div",Yr,[P("div",qr,[e[52]||(e[52]=P("div",{class:"ciphers-group__title"},"TLS 1.3",-1)),d(r(We),{value:u.value,"onUpdate:value":e[22]||(e[22]=i=>u.value=i)},{default:a(()=>[d(r(oe),null,{default:a(()=>[(b(!0),M(re,null,Ve(r(he),i=>(b(),L(r(Me),{key:i.value,value:i.value,label:i.label},null,8,["value","label"]))),128))]),_:1})]),_:1},8,["value"])]),P("div",Kr,[e[53]||(e[53]=P("div",{class:"ciphers-group__title"},"TLS 1.2",-1)),d(r(We),{value:u.value,"onUpdate:value":e[23]||(e[23]=i=>u.value=i)},{default:a(()=>[d(r(oe),null,{default:a(()=>[(b(!0),M(re,null,Ve(r(pe),i=>(b(),L(r(Me),{key:i.value,value:i.value,label:i.label},null,8,["value","label"]))),128))]),_:1})]),_:1},8,["value"])]),P("div",Xr,[e[54]||(e[54]=P("div",{class:"ciphers-group__title"},"国密 NTLS",-1)),d(r(We),{value:u.value,"onUpdate:value":e[24]||(e[24]=i=>u.value=i)},{default:a(()=>[d(r(oe),null,{default:a(()=>[(b(!0),M(re,null,Ve(r(me),i=>(b(),L(r(Me),{key:i.value,value:i.value,label:i.label},null,8,["value","label"]))),128))]),_:1})]),_:1},8,["value"])]),P("div",Jr,[e[55]||(e[55]=P("span",{class:"form-hint"},"已选：",-1)),P("span",Qr,A(u.value.length>0?u.value.join(":"):"（未选择，保存时使用默认套件）"),1)])])):(b(),L(r(je),{key:2,value:f.value,"onUpdate:value":e[25]||(e[25]=i=>f.value=i),type:"textarea",autosize:{minRows:2,maxRows:6},placeholder:r(Eo)(c.mode)},null,8,["value","placeholder"]))]),_:1})]),_:1})]),_:1},8,["model"]))],512),[[Oe,k.value==="ssl"]])]),_:1},8,["show"])]),_:1})]),_:1},8,["show"]))}}),el=Ho(Zr,[["__scopeId","data-v-fd042e20"]]),cl=Object.freeze(Object.defineProperty({__proto__:null,default:el},Symbol.toStringTag,{value:"Module"}));export{Me as C,el as P,Ye as R,tr as a,We as b,jo as c,dr as d,cl as e,Mt as r,or as s,Dr as u};
