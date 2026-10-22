import{a as oe,e as me,F as xe,h as Y,I as we,m as _e}from"./modules-7Z2VbH15.js";import{F as Ce,aS as q,H as V,aA as w,I as m,J as R,az as O,a9 as ze,T as J,b5 as te,ab as k,aa as Q,aR as E,_ as ye,ar as W,d as X,K as G,G as Re,o as T,c as H,a as K,O as M,L as I,R as ae,aE as re,V as ne,W as P,aI as Se,X as L,P as ie,bl as ke,ag as Be,a8 as Fe,w as _,b as y,u as C,e as U,B as ee,br as $e,a6 as Ie}from"./index-QkpsHkNI.js";var Ve={radioSizeSmall:"14px",radioSizeMedium:"16px",radioSizeLarge:"18px",labelPadding:"0 8px",labelFontWeight:"400"};function De(o){const{borderColor:e,primaryColor:r,baseColor:n,textColorDisabled:d,inputColorDisabled:p,textColor2:u,opacityDisabled:t,borderRadius:s,fontSizeSmall:c,fontSizeMedium:f,fontSizeLarge:v,heightSmall:i,heightMedium:a,heightLarge:h,lineHeight:z}=o;return{...Ve,labelLineHeight:z,buttonHeightSmall:i,buttonHeightMedium:a,buttonHeightLarge:h,fontSizeSmall:c,fontSizeMedium:f,fontSizeLarge:v,boxShadow:`inset 0 0 0 1px ${e}`,boxShadowActive:`inset 0 0 0 1px ${r}`,boxShadowFocus:`inset 0 0 0 1px ${r}, 0 0 0 2px ${q(r,{alpha:.2})}`,boxShadowHover:`inset 0 0 0 1px ${r}`,boxShadowDisabled:`inset 0 0 0 1px ${e}`,color:n,colorDisabled:p,colorActive:"#0000",textColor:u,textColorDisabled:d,dotColorActive:r,dotColorDisabled:e,buttonBorderColor:e,buttonBorderColorActive:r,buttonBorderColorHover:e,buttonColor:n,buttonColorActive:n,buttonTextColor:u,buttonTextColorActive:r,buttonTextColorHover:r,opacityDisabled:t,buttonBoxShadowFocus:`inset 0 0 0 1px ${r}, 0 0 0 2px ${q(r,{alpha:.3})}`,buttonBoxShadowHover:"inset 0 0 0 1px #0000",buttonBoxShadow:"inset 0 0 0 1px #0000",buttonBorderRadius:s}}const le={name:"Radio",common:Ce,self:De};var Ae=V("radio",`
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
`,[w("checked",[m("dot",`
 background-color: var(--n-color-active);
 `)]),m("dot-wrapper",`
 position: relative;
 flex-shrink: 0;
 flex-grow: 0;
 width: var(--n-radio-size);
 `),V("radio-input",`
 position: absolute;
 border: 0;
 width: 0;
 height: 0;
 opacity: 0;
 margin: 0;
 `),m("dot",`
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
 `,[R("&::before",`
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
 `),w("checked",{boxShadow:"var(--n-box-shadow-active)"},[R("&::before",`
 opacity: 1;
 transform: scale(1);
 `)])]),m("label",`
 color: var(--n-text-color);
 padding: var(--n-label-padding);
 font-weight: var(--n-label-font-weight);
 display: inline-block;
 transition: color .3s var(--n-bezier);
 `),O("disabled",`
 cursor: pointer;
 `,[R("&:hover",[m("dot",{boxShadow:"var(--n-box-shadow-hover)"})]),w("focus",[R("&:not(:active)",[m("dot",{boxShadow:"var(--n-box-shadow-focus)"})])])]),w("disabled",`
 cursor: not-allowed;
 `,[m("dot",{boxShadow:"var(--n-box-shadow-disabled)",backgroundColor:"var(--n-color-disabled)"},[R("&::before",{backgroundColor:"var(--n-dot-color-disabled)"}),w("checked",`
 opacity: 1;
 `)]),m("label",{color:"var(--n-text-color-disabled)"}),V("radio-input",`
 cursor: not-allowed;
 `)])]);const Te={name:String,value:{type:[String,Number,Boolean],default:"on"},checked:{type:Boolean,default:void 0},defaultChecked:Boolean,disabled:{type:Boolean,default:void 0},label:String,size:String,onUpdateChecked:[Function,Array],"onUpdate:checked":[Function,Array],checkedValue:{type:Boolean,default:void 0}},se=ye("n-radio-group");function Ee(o){const e=ze(se,null),{mergedClsPrefixRef:r,mergedComponentPropsRef:n}=J(o),d=te(o,{mergedSize(l){var x,D;const{size:b}=o;if(b!==void 0)return b;if(e){const{mergedSizeRef:{value:A}}=e;if(A!==void 0)return A}if(l)return l.mergedSize.value;const g=(D=(x=n==null?void 0:n.value)==null?void 0:x.Radio)==null?void 0:D.size;return g||"medium"},mergedDisabled(l){return!!(o.disabled||e!=null&&e.disabledRef.value||l!=null&&l.disabled.value)}}),{mergedSizeRef:p,mergedDisabledRef:u}=d,t=k(null),s=k(null),c=k(o.defaultChecked),f=W(o,"checked"),v=oe(f,c),i=Q(()=>e?e.valueRef.value===o.value:v.value),a=Q(()=>{const{name:l}=o;if(l!==void 0)return l;if(e)return e.nameRef.value}),h=k(!1);function z(){if(e){const{doUpdateValue:l}=e,{value:b}=o;E(l,b)}else{const{onUpdateChecked:l,"onUpdate:checked":b}=o,{nTriggerFormInput:g,nTriggerFormChange:x}=d;l&&E(l,!0),b&&E(b,!0),g(),x(),c.value=!0}}function B(){u.value||i.value||z()}function F(){B(),t.value&&(t.value.checked=i.value)}function $(){h.value=!1}function S(){h.value=!0}return{mergedClsPrefix:e?e.mergedClsPrefixRef:r,inputRef:t,labelRef:s,mergedName:a,mergedDisabled:u,renderSafeChecked:i,focus:h,mergedSize:p,handleRadioInputChange:F,handleRadioInputBlur:$,handleRadioInputFocus:S}}const Pe=["value","name","checked","disabled","onChange","onFocus","onBlur"],Ue={...G.props,...Te};var Z=X({name:"Radio",props:Ue,setup(o){const e=Ee(o),r=G("Radio","-radio",Ae,le,o,e.mergedClsPrefix),n=P(()=>{const{mergedSize:{value:c}}=e,{common:{cubicBezierEaseInOut:f},self:{boxShadow:v,boxShadowActive:i,boxShadowDisabled:a,boxShadowFocus:h,boxShadowHover:z,color:B,colorDisabled:F,colorActive:$,textColor:S,textColorDisabled:l,dotColorActive:b,dotColorDisabled:g,labelPadding:x,labelLineHeight:D,labelFontWeight:A,[L("fontSize",c)]:N,[L("radioSize",c)]:j}}=r.value;return{"--n-bezier":f,"--n-label-line-height":D,"--n-label-font-weight":A,"--n-box-shadow":v,"--n-box-shadow-active":i,"--n-box-shadow-disabled":a,"--n-box-shadow-focus":h,"--n-box-shadow-hover":z,"--n-color":B,"--n-color-active":$,"--n-color-disabled":F,"--n-dot-color-active":b,"--n-dot-color-disabled":g,"--n-font-size":N,"--n-radio-size":j,"--n-text-color":S,"--n-text-color-disabled":l,"--n-label-padding":x}}),{inlineThemeDisabled:d,mergedClsPrefixRef:p,mergedRtlRef:u}=J(o),t=re("Radio",u,p),s=d?ne("radio",P(()=>e.mergedSize.value[0]),n,o):void 0;return Object.assign(e,{rtlEnabled:t,cssVars:d?void 0:n,themeClass:s==null?void 0:s.themeClass,onRender:s==null?void 0:s.onRender})},render(){const{$slots:o,mergedClsPrefix:e,onRender:r,label:n}=this;return r==null||r(),(()=>{const d=Re("f8c6901d8cd45c02");return T(),H("label",{class:I([`${e}-radio`,this.themeClass,this.rtlEnabled&&`${e}-radio--rtl`,this.mergedDisabled&&`${e}-radio--disabled`,this.renderSafeChecked&&`${e}-radio--checked`,this.focus&&`${e}-radio--focus`]),style:ae(this.cssVars)},[K("div",{class:I(`${e}-radio__dot-wrapper`)},[d[0]||(d[0]=M(" ",-1)),K("div",{class:I([`${e}-radio__dot`,this.renderSafeChecked&&`${e}-radio__dot--checked`])},null,2),K("input",{ref:"inputRef",type:"radio",class:I(`${e}-radio-input`),value:this.value,name:this.mergedName,checked:this.renderSafeChecked,disabled:this.mergedDisabled,onChange:this.handleRadioInputChange,onFocus:this.handleRadioInputFocus,onBlur:this.handleRadioInputBlur},null,42,Pe)],2),M(()=>Se(o.default,p=>!p&&!n?null:(T(),H("div",{ref:"labelRef",class:I(`${e}-radio__label`)},[M(()=>p||n)],2))))],6)})()}});function Me(o,e="default",r=[]){const n=o.$slots[e];return n===void 0?r:n()}var He=V("radio-group",`
 display: inline-block;
 font-size: var(--n-font-size);
`,[m("splitor",`
 display: inline-block;
 vertical-align: bottom;
 width: 1px;
 transition:
 background-color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 background: var(--n-button-border-color);
 `,[w("checked",{backgroundColor:"var(--n-button-border-color-active)"}),w("disabled",{opacity:"var(--n-opacity-disabled)"})]),w("button-group",`
 white-space: nowrap;
 height: var(--n-height);
 line-height: var(--n-height);
 `,[V("radio-button",{height:"var(--n-height)",lineHeight:"var(--n-height)"}),m("splitor",{height:"var(--n-height)"})]),V("radio-button",`
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
 `,[V("radio-input",`
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
 `),m("state-border",`
 z-index: 1;
 pointer-events: none;
 position: absolute;
 box-shadow: var(--n-button-box-shadow);
 transition: box-shadow .3s var(--n-bezier);
 left: -1px;
 bottom: -1px;
 right: -1px;
 top: -1px;
 `),R("&:first-child",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 border-left: 1px solid var(--n-button-border-color);
 `,[m("state-border",`
 border-top-left-radius: var(--n-button-border-radius);
 border-bottom-left-radius: var(--n-button-border-radius);
 `)]),R("&:last-child",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 border-right: 1px solid var(--n-button-border-color);
 `,[m("state-border",`
 border-top-right-radius: var(--n-button-border-radius);
 border-bottom-right-radius: var(--n-button-border-radius);
 `)]),O("disabled",`
 cursor: pointer;
 `,[R("&:hover",[m("state-border",`
 transition: box-shadow .3s var(--n-bezier);
 box-shadow: var(--n-button-box-shadow-hover);
 `),O("checked",{color:"var(--n-button-text-color-hover)"})]),w("focus",[R("&:not(:active)",[m("state-border",{boxShadow:"var(--n-button-box-shadow-focus)"})])])]),w("checked",`
 background: var(--n-button-color-active);
 color: var(--n-button-text-color-active);
 border-color: var(--n-button-border-color-active);
 `),w("disabled",`
 cursor: not-allowed;
 opacity: var(--n-opacity-disabled);
 `)])]);const Le=["onFocusin","onFocusout"];function Ge(o,e,r){var p;const n=[];let d=!1;for(let u=0;u<o.length;++u){const t=o[u],s=(p=t.type)==null?void 0:p.name;s==="RadioButton"&&(d=!0);const c=t.props;if(s!=="RadioButton"){n.push(t);continue}if(u===0)n.push(t);else{const f=n[n.length-1].props,v=e===f.value,i=f.disabled,a=e===c.value,h=c.disabled,z=(v?2:0)+(i?0:1),B=(a?2:0)+(h?0:1),F={[`${r}-radio-group__splitor--disabled`]:i,[`${r}-radio-group__splitor--checked`]:v},$={[`${r}-radio-group__splitor--disabled`]:h,[`${r}-radio-group__splitor--checked`]:a},S=z<B?$:F;n.push((T(),H("div",{key:1,class:I([`${r}-radio-group__splitor`,S])},null,2)),t)}}return{children:n,isButtonGroup:d}}const Ne={...G.props,name:String,options:Array,labelField:{type:String,default:"label"},valueField:{type:String,default:"value"},value:[String,Number,Boolean],defaultValue:{type:[String,Number,Boolean],default:null},size:String,disabled:{type:Boolean,default:void 0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array]};var je=X({name:"RadioGroup",props:Ne,setup(o){const e=k(null),{mergedSizeRef:r,mergedDisabledRef:n,nTriggerFormChange:d,nTriggerFormInput:p,nTriggerFormBlur:u,nTriggerFormFocus:t}=te(o),{mergedClsPrefixRef:s,inlineThemeDisabled:c,mergedRtlRef:f}=J(o),v=G("Radio","-radio-group",He,le,o,s),i=k(o.defaultValue),a=W(o,"value"),h=oe(a,i);function z(b){const{onUpdateValue:g,"onUpdate:value":x}=o;g&&E(g,b),x&&E(x,b),i.value=b,d(),p()}function B(b){const{value:g}=e;g&&(g.contains(b.relatedTarget)||t())}function F(b){const{value:g}=e;g&&(g.contains(b.relatedTarget)||u())}Be(se,{mergedClsPrefixRef:s,nameRef:W(o,"name"),valueRef:h,disabledRef:n,mergedSizeRef:r,doUpdateValue:z});const $=re("Radio",f,s),S=P(()=>{const{value:b}=r,{common:{cubicBezierEaseInOut:g},self:{buttonBorderColor:x,buttonBorderColorActive:D,buttonBorderRadius:A,buttonBoxShadow:N,buttonBoxShadowFocus:j,buttonBoxShadowHover:de,buttonColor:ue,buttonColorActive:ce,buttonTextColor:be,buttonTextColorActive:he,buttonTextColorHover:pe,opacityDisabled:fe,[L("buttonHeight",b)]:ve,[L("fontSize",b)]:ge}}=v.value;return{"--n-font-size":ge,"--n-bezier":g,"--n-button-border-color":x,"--n-button-border-color-active":D,"--n-button-border-radius":A,"--n-button-box-shadow":N,"--n-button-box-shadow-focus":j,"--n-button-box-shadow-hover":de,"--n-button-color":ue,"--n-button-color-active":ce,"--n-button-text-color":be,"--n-button-text-color-hover":pe,"--n-button-text-color-active":he,"--n-height":ve,"--n-opacity-disabled":fe}}),l=c?ne("radio-group",P(()=>r.value[0]),S,o):void 0;return{selfElRef:e,rtlEnabled:$,mergedClsPrefix:s,mergedValue:h,handleFocusout:F,handleFocusin:B,cssVars:c?void 0:S,themeClass:l==null?void 0:l.themeClass,onRender:l==null?void 0:l.onRender}},render(){var c;const{mergedValue:o,mergedClsPrefix:e,handleFocusin:r,handleFocusout:n}=this,{options:d,labelField:p,valueField:u}=this.$props,{children:t,isButtonGroup:s}=Ge(d?d.map(f=>{const v=f[u];return T(),ie(Z,{key:typeof v=="boolean"?`__n_${v}`:v,value:v,disabled:f.disabled,label:f[p]},null,8,["value","disabled","label"])}):ke(Me(this)),o,e);return(c=this.onRender)==null||c.call(this),T(),H("div",{onFocusin:r,onFocusout:n,ref:"selfElRef",class:I([`${e}-radio-group`,this.rtlEnabled&&`${e}-radio-group--rtl`,this.themeClass,s&&`${e}-radio-group--button-group`]),style:ae(this.cssVars)},[M(()=>t)],46,Le)}});const We=X({__name:"InstallModal",props:{visible:{type:Boolean}},emits:["update:visible","success"],setup(o,{emit:e}){const r=o,n=e,d=me(),p=k(null),u=k(!1),t=Ie({install_type:"zip",file_path:"",source_path:""});Fe(()=>t.install_type,i=>{i==="zip"?t.source_path="":t.file_path=""});const s={file_path:[{validator:(i,a)=>t.install_type==="zip"&&!a?new Error("请输入 ZIP 文件路径"):!0,trigger:"blur"}],source_path:[{validator:(i,a)=>t.install_type==="path"&&!a?new Error("请输入源码目录路径"):!0,trigger:"blur"}]},c=P({get:()=>r.visible,set:i=>n("update:visible",i)});function f(){c.value=!1}async function v(){var i;try{await((i=p.value)==null?void 0:i.validate()),u.value=!0;const a={install_type:t.install_type};t.install_type==="zip"?a.file_path=t.file_path:a.source_path=t.source_path,await _e.install(a),d.success("安装成功"),n("success"),f()}catch(a){d.error(a.message||"安装失败")}finally{u.value=!1}}return(i,a)=>(T(),ie(C($e),{show:c.value,"onUpdate:show":a[2]||(a[2]=h=>c.value=h),title:"安装模块",preset:"dialog",style:{width:"500px"},onClose:f},{action:_(()=>[y(C(ee),{onClick:f},{default:_(()=>[...a[5]||(a[5]=[U("取消",-1)])]),_:1}),y(C(ee),{type:"primary",loading:u.value,onClick:v},{default:_(()=>[...a[6]||(a[6]=[U("确定",-1)])]),_:1},8,["loading"])]),default:_(()=>[y(C(xe),{ref_key:"formRef",ref:p,model:t,rules:s,"label-placement":"left","label-width":"120"},{default:_(()=>[y(C(Y),{label:"安装方式",path:"install_type"},{default:_(()=>[y(C(je),{value:t.install_type,"onUpdate:value":a[0]||(a[0]=h=>t.install_type=h)},{default:_(()=>[y(C(Z),{value:"zip"},{default:_(()=>[...a[3]||(a[3]=[U("ZIP 包路径",-1)])]),_:1}),y(C(Z),{value:"path"},{default:_(()=>[...a[4]||(a[4]=[U("源码目录路径",-1)])]),_:1})]),_:1},8,["value"])]),_:1}),y(C(Y),{label:t.install_type==="zip"?"ZIP 文件路径":"源码目录路径",path:t.install_type==="zip"?"file_path":"source_path"},{default:_(()=>[y(C(we),{value:t[t.install_type==="zip"?"file_path":"source_path"],"onUpdate:value":a[1]||(a[1]=h=>t[t.install_type==="zip"?"file_path":"source_path"]=h),placeholder:t.install_type==="zip"?"例如 /tmp/module.zip":"例如 /opt/modules/my-module"},null,8,["value","placeholder"])]),_:1},8,["label","path"])]),_:1},8,["model"])]),_:1},8,["show"]))}});export{Z as R,We as _,je as a,Me as g,le as r};
