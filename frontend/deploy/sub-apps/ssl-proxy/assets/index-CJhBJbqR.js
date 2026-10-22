import{x as Yt,y as Te,z as qt,d as se,A as Jt,C as Zt,D as O,g as Qt,h as Z,F as ea,G as ta,f as aa,H as ra,o as l,c as y,n as S,m as ie,l as E,J as lt,K as na,L as _e,M as Y,s as k,I as st,N as oa,O as ia,r as le,P as la,i as r,Q as $,R as d,j as D,T as sa,B as pe,k as dt,U as $e,V as Be,p as da,X as ca,Y as ve,Z as ct,_ as ba,q as fa,$ as Ke,a0 as ua,a1 as Ye,a2 as ge,t as te,a3 as Ce,a4 as we,a5 as pa,a6 as va,a7 as ga,a8 as ae,a9 as Se,a as B,w as x,u as v,aa as qe,ab as G,e as U,ac as oe,ad as ha,ae as ma,b as xa,af as ya}from"./index-CVUczFfF.js";import{T as Ca}from"./Tag-DPmVnj_2.js";import{D as K,a as Sa}from"./DescriptionsItem-31qlb8JN.js";import{u as wa}from"./use-message-D3mBhG8n.js";import{u as Je,S as Ta,a as Ae}from"./Spin-QgI_6sok.js";import{A as _a,u as Pa,p as Ze}from"./proxy-BHDJblh3.js";import{C as ka,u as Ra}from"./useProxyAction-B-0pBidt.js";import{c as za,a as Qe,o as La,P as $a,S as Ba}from"./proxyOptions-BqbSonRo.js";import et from"./ConfigTab-CTd6R6V_.js";import Aa from"./LogsTab-D6ix2wA7.js";import{_ as Wa}from"./_plugin-vue_export-helper-DlAUqK2U.js";var Ea=/\s/;function Ia(e){for(var n=e.length;n--&&Ea.test(e.charAt(n)););return n}var Va=/^\s+/;function Ma(e){return e&&e.slice(0,Ia(e)+1).replace(Va,"")}var tt=NaN,Oa=/^[-+]0x[0-9a-f]+$/i,Ha=/^0b[01]+$/i,ja=/^0o[0-7]+$/i,Da=parseInt;function at(e){if(typeof e=="number")return e;if(Yt(e))return tt;if(Te(e)){var n=typeof e.valueOf=="function"?e.valueOf():e;e=Te(n)?n+"":n}if(typeof e!="string")return e===0?e:+e;e=Ma(e);var i=Ha.test(e);return i||ja.test(e)?Da(e.slice(2),i?2:8):Oa.test(e)?tt:+e}var We=function(){return qt.Date.now()},Ga="Expected a function",Fa=Math.max,Na=Math.min;function Ua(e,n,i){var u,f,R,p,b,C,g=0,I=!1,s=!1,z=!0;if(typeof e!="function")throw new TypeError(Ga);n=at(n)||0,Te(i)&&(I=!!i.leading,s="maxWait"in i,R=s?Fa(at(i.maxWait)||0,n):R,z="trailing"in i?!!i.trailing:z);function _(h){var m=u,P=f;return u=f=void 0,g=h,p=e.apply(P,m),p}function w(h){return g=h,b=setTimeout(H,n),I?_(h):p}function T(h){var m=h-C,P=h-g,j=n-m;return s?Na(j,R-P):j}function M(h){var m=h-C,P=h-g;return C===void 0||m>=n||m<0||s&&P>=R}function H(){var h=We();if(M(h))return Q(h);b=setTimeout(H,T(h))}function Q(h){return b=void 0,z&&u?_(h):(u=f=void 0,p)}function re(){b!==void 0&&clearTimeout(b),g=0,u=C=f=b=void 0}function J(){return b===void 0?p:Q(We())}function q(){var h=We(),m=M(h);if(u=arguments,f=this,C=h,m){if(b===void 0)return w(C);if(s)return clearTimeout(b),b=setTimeout(H,n),_(C)}return b===void 0&&(b=setTimeout(H,n)),p}return q.cancel=re,q.flush=J,q}var Xa="Expected a function";function Ka(e,n,i){var u=!0,f=!0;if(typeof e!="function")throw new TypeError(Xa);return Te(i)&&(u="leading"in i?!!i.leading:u,f="trailing"in i?!!i.trailing:f),Ua(e,n,{leading:u,maxWait:n,trailing:f})}const Ya=Qe(".v-x-scroll",{overflow:"auto",scrollbarWidth:"none"},[Qe("&::-webkit-scrollbar",{width:0,height:0})]),qa=se({name:"XScroll",props:{disabled:Boolean,onScroll:Function},setup(){const e=O(null);function n(f){!(f.currentTarget.offsetWidth<f.currentTarget.scrollWidth)||f.deltaY===0||(f.currentTarget.scrollLeft+=f.deltaY+f.deltaX,f.preventDefault())}const i=Zt();return Ya.mount({id:"vueuc/x-scroll",head:!0,anchorMetaName:za,ssr:i}),Object.assign({selfRef:e,handleWheel:n},{scrollTo(...f){var R;(R=e.value)===null||R===void 0||R.scrollTo(...f)}})},render(){return Jt("div",{ref:"selfRef",onScroll:this.onScroll,onWheel:this.disabled?void 0:this.handleWheel,class:"v-x-scroll"},this.$slots)}});var Ja=se({name:"ChevronLeft",render(){return(()=>{const e=Qt("dfe229c2639b2082");return e[0]||(e[0]=Z("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},[Z("path",{d:"M10.3536 3.14645C10.5488 3.34171 10.5488 3.65829 10.3536 3.85355L6.20711 8L10.3536 12.1464C10.5488 12.3417 10.5488 12.6583 10.3536 12.8536C10.1583 13.0488 9.84171 13.0488 9.64645 12.8536L5.14645 8.35355C4.95118 8.15829 4.95118 7.84171 5.14645 7.64645L9.64645 3.14645C9.84171 2.95118 10.1583 2.95118 10.3536 3.14645Z",fill:"currentColor"})],-1))})()}}),Za={tabFontSizeSmall:"14px",tabFontSizeMedium:"14px",tabFontSizeLarge:"16px",tabGapSmallLine:"36px",tabGapMediumLine:"36px",tabGapLargeLine:"36px",tabGapSmallLineVertical:"8px",tabGapMediumLineVertical:"8px",tabGapLargeLineVertical:"8px",tabPaddingSmallLine:"6px 0",tabPaddingMediumLine:"10px 0",tabPaddingLargeLine:"14px 0",tabPaddingVerticalSmallLine:"6px 12px",tabPaddingVerticalMediumLine:"8px 16px",tabPaddingVerticalLargeLine:"10px 20px",tabGapSmallBar:"36px",tabGapMediumBar:"36px",tabGapLargeBar:"36px",tabGapSmallBarVertical:"8px",tabGapMediumBarVertical:"8px",tabGapLargeBarVertical:"8px",tabPaddingSmallBar:"4px 0",tabPaddingMediumBar:"6px 0",tabPaddingLargeBar:"10px 0",tabPaddingVerticalSmallBar:"6px 12px",tabPaddingVerticalMediumBar:"8px 16px",tabPaddingVerticalLargeBar:"10px 20px",tabGapSmallCard:"4px",tabGapMediumCard:"4px",tabGapLargeCard:"4px",tabGapSmallCardVertical:"4px",tabGapMediumCardVertical:"4px",tabGapLargeCardVertical:"4px",tabPaddingSmallCard:"8px 16px",tabPaddingMediumCard:"10px 20px",tabPaddingLargeCard:"12px 24px",tabPaddingSmallSegment:"4px 0",tabPaddingMediumSegment:"6px 0",tabPaddingLargeSegment:"8px 0",tabPaddingVerticalLargeSegment:"0 8px",tabPaddingVerticalSmallCard:"8px 12px",tabPaddingVerticalMediumCard:"10px 16px",tabPaddingVerticalLargeCard:"12px 20px",tabPaddingVerticalSmallSegment:"0 4px",tabPaddingVerticalMediumSegment:"0 6px",tabGapSmallSegment:"0",tabGapMediumSegment:"0",tabGapLargeSegment:"0",tabGapSmallSegmentVertical:"0",tabGapMediumSegmentVertical:"0",tabGapLargeSegmentVertical:"0",panePaddingSmall:"8px 0 0 0",panePaddingMedium:"12px 0 0 0",panePaddingLarge:"16px 0 0 0",closeSize:"18px",closeIconSize:"14px"};function Qa(e){const{textColor2:n,primaryColor:i,textColorDisabled:u,closeIconColor:f,closeIconColorHover:R,closeIconColorPressed:p,closeColorHover:b,closeColorPressed:C,tabColor:g,baseColor:I,dividerColor:s,fontWeight:z,textColor1:_,borderRadius:w,fontSize:T,fontWeightStrong:M}=e;return{...Za,colorSegment:g,tabFontSizeCard:T,tabTextColorLine:_,tabTextColorActiveLine:i,tabTextColorHoverLine:i,tabTextColorDisabledLine:u,tabTextColorSegment:_,tabTextColorActiveSegment:n,tabTextColorHoverSegment:n,tabTextColorDisabledSegment:u,tabTextColorBar:_,tabTextColorActiveBar:i,tabTextColorHoverBar:i,tabTextColorDisabledBar:u,tabTextColorCard:_,tabTextColorHoverCard:_,tabTextColorActiveCard:i,tabTextColorDisabledCard:u,barColor:i,closeIconColor:f,closeIconColorHover:R,closeIconColorPressed:p,closeColorHover:b,closeColorPressed:C,closeBorderRadius:w,tabColor:g,tabColorSegment:I,tabBorderColor:s,tabFontWeightActive:z,tabFontWeight:z,tabBorderRadius:w,paneTextColor:n,fontWeightStrong:M}}const er=ea({name:"Tabs",common:aa,peers:{Button:ta},self:Qa}),Oe=ra("n-tabs"),bt={tab:[String,Number,Object,Function],name:{type:[String,Number],required:!0},disabled:Boolean,displayDirective:{type:String,default:"if"},closable:{type:Boolean,default:void 0},tabProps:Object,label:[String,Number,Object,Function]};var Ee=se({__TAB_PANE__:!0,name:"TabPane",alias:["TabPanel"],props:bt,slots:Object,setup(e){const n=lt(Oe,null);return n||na("tab-pane","`n-tab-pane` must be placed inside `n-tabs`."),{style:n.paneStyleRef,class:n.paneClassRef,mergedClsPrefix:n.mergedClsPrefixRef}},render(){return l(),y("div",{class:E([`${this.mergedClsPrefix}-tab-pane`,this.class]),style:ie(this.style)},[S(()=>{var e,n;return(n=(e=this.$slots).default)==null?void 0:n.call(e)})],6)}});const tr=["data-name","data-disabled"],ar={internalLeftPadded:Boolean,internalAddable:Boolean,internalCreatedByPane:Boolean,...la(bt,["displayDirective"])};var Me=se({__TAB__:!0,inheritAttrs:!1,name:"Tab",props:ar,setup(e){const{mergedClsPrefixRef:n,valueRef:i,typeRef:u,closableRef:f,tabStyleRef:R,addTabStyleRef:p,tabClassRef:b,addTabClassRef:C,tabChangeIdRef:g,onBeforeLeaveRef:I,triggerRef:s,handleAdd:z,activateTab:_,handleClose:w}=lt(Oe);return{trigger:s,mergedClosable:le(()=>{if(e.internalAddable)return!1;const{closable:T}=e;return T===void 0?f.value:T}),style:R,addStyle:p,tabClass:b,addTabClass:C,clsPrefix:n,value:i,type:u,handleClose(T){T.stopPropagation(),!e.disabled&&w(e.name)},activateTab(){if(e.disabled)return;if(e.internalAddable){z();return}const{name:T}=e,M=++g.id;if(T!==i.value){const{value:H}=I;H?Promise.resolve(H(e.name,i.value)).then(Q=>{Q&&g.id===M&&_(T)}):_(T)}}}},render(){const{internalAddable:e,clsPrefix:n,name:i,disabled:u,label:f,tab:R,value:p,mergedClosable:b,trigger:C,$slots:{default:g}}=this,I=f??R;return l(),y("div",{class:E(`${n}-tabs-tab-wrapper`)},[this.internalLeftPadded?(l(),y("div",{key:0,class:E(`${n}-tabs-tab-pad`)},null,2)):S(()=>null),(l(),y("div",_e({key:i,"data-name":i,"data-disabled":u?!0:void 0},_e({class:[`${n}-tabs-tab`,p===i&&`${n}-tabs-tab--active`,u&&`${n}-tabs-tab--disabled`,b&&`${n}-tabs-tab--closable`,e&&`${n}-tabs-tab--addable`,e?this.addTabClass:this.tabClass],onClick:C==="click"?this.activateTab:void 0,onMouseenter:C==="hover"?this.activateTab:void 0,style:e?this.addStyle:this.style},this.internalCreatedByPane?this.tabProps||{}:this.$attrs)),[Z("span",{class:E(`${n}-tabs-tab__label`)},[e?(l(),y(Y,{key:0},[Z("div",{class:E(`${n}-tabs-tab__height-placeholder`)}," ",2),(l(),k(st,{clsPrefix:n},{default:()=>(l(),k(_a))},1032,["clsPrefix"]))],64)):(l(),y(Y,{key:1},[g?(l(),y(Y,{key:0},[S(()=>g())],64)):(l(),y(Y,{key:1},[typeof I=="object"?(l(),y(Y,{key:0},[S(()=>I)],64)):(l(),y(Y,{key:1},[S(()=>oa(I??i))],64))],64))],64))],2),b&&this.type==="card"?(l(),k(ia,{key:0,clsPrefix:n,class:E(`${n}-tabs-tab__close`),onClick:this.handleClose,disabled:u},null,8,["clsPrefix","class","onClick","disabled"])):S(()=>null)],16,tr))],2)}}),rr=r("tabs",`
 box-sizing: border-box;
 width: 100%;
 display: flex;
 flex-direction: column;
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
`,[$("&.transition-disabled",[r("tabs-tab",`
 transition: none !important;
 `),r("tabs-nav-scroll-content",`
 transition: none !important;
 `),r("tabs-tab-pad",`
 transition: none !important;
 `)]),d("segment-type",[r("tabs-rail",[$("&.transition-disabled",[r("tabs-capsule",`
 transition: none;
 `)])])]),d("top",[r("tab-pane",`
 padding: var(--n-pane-padding-top) var(--n-pane-padding-right) var(--n-pane-padding-bottom) var(--n-pane-padding-left);
 `)]),d("left",[r("tab-pane",`
 padding: var(--n-pane-padding-right) var(--n-pane-padding-bottom) var(--n-pane-padding-left) var(--n-pane-padding-top);
 `)]),d("left, right",`
 flex-direction: row;
 `,[r("tabs-bar",`
 width: 2px;
 right: 0;
 transition:
 top .2s var(--n-bezier),
 max-height .2s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `),r("tabs-tab",`
 padding: var(--n-tab-padding-vertical); 
 `)]),d("right",`
 flex-direction: row-reverse;
 `,[r("tab-pane",`
 padding: var(--n-pane-padding-left) var(--n-pane-padding-top) var(--n-pane-padding-right) var(--n-pane-padding-bottom);
 `),r("tabs-bar",`
 left: 0;
 `)]),d("bottom",`
 flex-direction: column-reverse;
 justify-content: flex-end;
 `,[r("tab-pane",`
 padding: var(--n-pane-padding-bottom) var(--n-pane-padding-right) var(--n-pane-padding-top) var(--n-pane-padding-left);
 `),r("tabs-bar",`
 top: 0;
 `)]),r("tabs-rail",`
 position: relative;
 padding: 3px;
 border-radius: var(--n-tab-border-radius);
 width: 100%;
 background-color: var(--n-color-segment);
 transition: background-color .3s var(--n-bezier);
 display: flex;
 align-items: center;
 `,[r("tabs-capsule",`
 border-radius: var(--n-tab-border-radius);
 position: absolute;
 left: 0;
 top: 0;
 pointer-events: none;
 background-color: var(--n-tab-color-segment);
 box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .08);
 transition: transform 0.3s var(--n-bezier);
 `),r("tabs-tab-wrapper",`
 flex-basis: 0;
 flex-grow: 1;
 display: flex;
 align-items: center;
 justify-content: center;
 `,[r("tabs-tab",`
 overflow: hidden;
 border-radius: var(--n-tab-border-radius);
 width: 100%;
 display: flex;
 align-items: center;
 justify-content: center;
 `,[d("active",`
 font-weight: var(--n-font-weight-strong);
 color: var(--n-tab-text-color-active);
 `),$("&:hover",`
 color: var(--n-tab-text-color-hover);
 `)])])]),d("flex",[r("tabs-nav",`
 width: 100%;
 position: relative;
 `,[r("tabs-wrapper",`
 width: 100%;
 `,[r("tabs-tab",`
 margin-right: 0;
 `)])])]),r("tabs-nav",`
 box-sizing: border-box;
 line-height: 1.5;
 display: flex;
 transition: border-color .3s var(--n-bezier);
 `,[D("prefix, suffix",`
 display: flex;
 align-items: center;
 `),D("prefix","padding-right: 16px;"),D("suffix","padding-left: 16px;")]),d("top, bottom",[$(">",[r("tabs-nav",[r("tabs-nav-scroll-wrapper",[$("&::before",`
 top: 0;
 bottom: 0;
 left: 0;
 width: 20px;
 `),$("&::after",`
 top: 0;
 bottom: 0;
 right: 0;
 width: 20px;
 `),d("shadow-start",[$("&::before",`
 box-shadow: inset 10px 0 8px -8px rgba(0, 0, 0, .12);
 `)]),d("shadow-end",[$("&::after",`
 box-shadow: inset -10px 0 8px -8px rgba(0, 0, 0, .12);
 `)])])])])]),d("left, right",[r("tabs-nav-scroll-content",`
 flex-direction: column;
 `),$(">",[r("tabs-nav",[r("tabs-nav-scroll-wrapper",[$("&::before",`
 top: 0;
 left: 0;
 right: 0;
 height: 20px;
 `),$("&::after",`
 bottom: 0;
 left: 0;
 right: 0;
 height: 20px;
 `),d("shadow-start",[$("&::before",`
 box-shadow: inset 0 10px 8px -8px rgba(0, 0, 0, .12);
 `)]),d("shadow-end",[$("&::after",`
 box-shadow: inset 0 -10px 8px -8px rgba(0, 0, 0, .12);
 `)])])])])]),r("tabs-nav-scroll-wrapper",`
 flex: 1;
 position: relative;
 overflow: hidden;
 `,[r("tabs-nav-y-scroll",`
 height: 100%;
 width: 100%;
 overflow-y: auto; 
 scrollbar-width: none;
 `,[$("&::-webkit-scrollbar, &::-webkit-scrollbar-track-piece, &::-webkit-scrollbar-thumb",`
 width: 0;
 height: 0;
 display: none;
 `)]),$("&::before, &::after",`
 transition: box-shadow .3s var(--n-bezier);
 pointer-events: none;
 content: "";
 position: absolute;
 z-index: 1;
 `),$("&.transition-disabled",[$("&::before, &::after",`
 transition: none;
 `)])]),r("tabs-nav-scroll-content",`
 display: flex;
 position: relative;
 min-width: 100%;
 min-height: 100%;
 width: fit-content;
 box-sizing: border-box;
 `),r("tabs-wrapper",`
 display: inline-flex;
 flex-wrap: nowrap;
 position: relative;
 `),r("tabs-tab-wrapper",`
 display: flex;
 flex-wrap: nowrap;
 flex-shrink: 0;
 flex-grow: 0;
 `),r("tabs-tab",`
 cursor: pointer;
 white-space: nowrap;
 flex-wrap: nowrap;
 display: inline-flex;
 align-items: center;
 color: var(--n-tab-text-color);
 font-size: var(--n-tab-font-size);
 background-clip: padding-box;
 padding: var(--n-tab-padding);
 transition:
 box-shadow .3s var(--n-bezier),
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[d("disabled",{cursor:"not-allowed"}),D("close",`
 margin-inline-start: 6px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `),D("label",`
 display: flex;
 align-items: center;
 z-index: 1;
 `)]),r("tabs-bar",`
 position: absolute;
 bottom: 0;
 height: 2px;
 border-radius: 1px;
 background-color: var(--n-bar-color);
 transition:
 left .2s var(--n-bezier),
 max-width .2s var(--n-bezier),
 opacity .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `,[$("&.transition-disabled",`
 transition: none;
 `),d("disabled",`
 background-color: var(--n-tab-text-color-disabled)
 `)]),r("tabs-pane-wrapper",`
 position: relative;
 overflow: hidden;
 transition: max-height .2s var(--n-bezier);
 `),r("tab-pane",`
 color: var(--n-pane-text-color);
 width: 100%;
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 opacity .2s var(--n-bezier);
 left: 0;
 right: 0;
 top: 0;
 `,[$("&.next-transition-leave-active, &.prev-transition-leave-active, &.next-transition-enter-active, &.prev-transition-enter-active",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 transform .2s var(--n-bezier),
 opacity .2s var(--n-bezier);
 `),$("&.next-transition-leave-active, &.prev-transition-leave-active",`
 position: absolute;
 `),$("&.next-transition-enter-from, &.prev-transition-leave-to",`
 transform: translateX(32px);
 opacity: 0;
 `),$("&.next-transition-leave-to, &.prev-transition-enter-from",`
 transform: translateX(-32px);
 opacity: 0;
 `),$("&.next-transition-leave-from, &.next-transition-enter-to, &.prev-transition-leave-from, &.prev-transition-enter-to",`
 transform: translateX(0);
 opacity: 1;
 `)]),r("tabs-tab-pad",`
 box-sizing: border-box;
 width: var(--n-tab-gap);
 flex-grow: 0;
 flex-shrink: 0;
 `),d("line-type, bar-type",[r("tabs-tab",`
 font-weight: var(--n-tab-font-weight);
 box-sizing: border-box;
 vertical-align: bottom;
 `,[$("&:hover",{color:"var(--n-tab-text-color-hover)"}),d("active",`
 color: var(--n-tab-text-color-active);
 font-weight: var(--n-tab-font-weight-active);
 `),d("disabled",{color:"var(--n-tab-text-color-disabled)"})])]),r("tabs-nav",[D("prefix, suffix",`
 border-color: var(--n-tab-border-color);
 `),r("tabs-nav-scroll-content",`
 border-color: var(--n-tab-border-color);
 `),d("line-type",[d("top",[D("prefix, suffix",`
 border-bottom: 1px solid var(--n-tab-border-color);
 `),r("tabs-nav-scroll-content",`
 border-bottom: 1px solid var(--n-tab-border-color);
 `),r("tabs-bar",`
 bottom: -1px;
 `)]),d("left",[D("prefix, suffix",`
 border-right: 1px solid var(--n-tab-border-color);
 `),r("tabs-nav-scroll-content",`
 border-right: 1px solid var(--n-tab-border-color);
 `),r("tabs-bar",`
 right: -1px;
 `)]),d("right",[D("prefix, suffix",`
 border-left: 1px solid var(--n-tab-border-color);
 `),r("tabs-nav-scroll-content",`
 border-left: 1px solid var(--n-tab-border-color);
 `),r("tabs-bar",`
 left: -1px;
 `)]),d("bottom",[D("prefix, suffix",`
 border-top: 1px solid var(--n-tab-border-color);
 `),r("tabs-nav-scroll-content",`
 border-top: 1px solid var(--n-tab-border-color);
 `),r("tabs-bar",`
 top: -1px;
 `)]),D("prefix, suffix",`
 transition: border-color .3s var(--n-bezier);
 `),r("tabs-nav-scroll-content",`
 transition: border-color .3s var(--n-bezier);
 `),r("tabs-bar",`
 border-radius: 0;
 `)]),d("card-type",[D("prefix, suffix",`
 transition: border-color .3s var(--n-bezier);
 `),r("tabs-pad",`
 flex-grow: 1;
 transition: border-color .3s var(--n-bezier);
 `),r("tabs-tab-pad",`
 transition: border-color .3s var(--n-bezier);
 `),r("tabs-tab",`
 font-weight: var(--n-tab-font-weight);
 border: 1px solid var(--n-tab-border-color);
 background-color: var(--n-tab-color);
 box-sizing: border-box;
 position: relative;
 vertical-align: bottom;
 display: flex;
 justify-content: space-between;
 font-size: var(--n-tab-font-size);
 color: var(--n-tab-text-color);
 `,[d("addable",`
 padding-left: 8px;
 padding-right: 8px;
 font-size: 16px;
 justify-content: center;
 `,[D("height-placeholder",`
 width: 0;
 font-size: var(--n-tab-font-size);
 `),sa("disabled",[$("&:hover",`
 color: var(--n-tab-text-color-hover);
 `)])]),d("closable","padding-inline-end: 8px;"),d("active",`
 background-color: #0000;
 font-weight: var(--n-tab-font-weight-active);
 color: var(--n-tab-text-color-active);
 `),d("disabled","color: var(--n-tab-text-color-disabled);")])]),d("left, right",`
 flex-direction: column; 
 `,[D("prefix, suffix",`
 padding: var(--n-tab-padding-vertical);
 `),r("tabs-wrapper",`
 flex-direction: column;
 `),r("tabs-tab-wrapper",`
 flex-direction: column;
 `,[r("tabs-tab-pad",`
 height: var(--n-tab-gap-vertical);
 width: 100%;
 `)])]),d("top",[d("card-type",[r("tabs-scroll-padding","border-bottom: 1px solid var(--n-tab-border-color);"),D("prefix, suffix",`
 border-bottom: 1px solid var(--n-tab-border-color);
 `),r("tabs-tab",`
 border-top-left-radius: var(--n-tab-border-radius);
 border-top-right-radius: var(--n-tab-border-radius);
 `,[d("active",`
 border-bottom: 1px solid #0000;
 `)]),r("tabs-tab-pad",`
 border-bottom: 1px solid var(--n-tab-border-color);
 `),r("tabs-pad",`
 border-bottom: 1px solid var(--n-tab-border-color);
 `)])]),d("left",[d("card-type",[r("tabs-scroll-padding","border-right: 1px solid var(--n-tab-border-color);"),D("prefix, suffix",`
 border-right: 1px solid var(--n-tab-border-color);
 `),r("tabs-tab",`
 border-top-left-radius: var(--n-tab-border-radius);
 border-bottom-left-radius: var(--n-tab-border-radius);
 `,[d("active",`
 border-right: 1px solid #0000;
 `)]),r("tabs-tab-pad",`
 border-right: 1px solid var(--n-tab-border-color);
 `),r("tabs-pad",`
 border-right: 1px solid var(--n-tab-border-color);
 `)])]),d("right",[d("card-type",[r("tabs-scroll-padding","border-left: 1px solid var(--n-tab-border-color);"),D("prefix, suffix",`
 border-left: 1px solid var(--n-tab-border-color);
 `),r("tabs-tab",`
 border-top-right-radius: var(--n-tab-border-radius);
 border-bottom-right-radius: var(--n-tab-border-radius);
 `,[d("active",`
 border-left: 1px solid #0000;
 `)]),r("tabs-tab-pad",`
 border-left: 1px solid var(--n-tab-border-color);
 `),r("tabs-pad",`
 border-left: 1px solid var(--n-tab-border-color);
 `)])]),d("bottom",[d("card-type",[r("tabs-scroll-padding","border-top: 1px solid var(--n-tab-border-color);"),D("prefix, suffix",`
 border-top: 1px solid var(--n-tab-border-color);
 `),r("tabs-tab",`
 border-bottom-left-radius: var(--n-tab-border-radius);
 border-bottom-right-radius: var(--n-tab-border-radius);
 `,[d("active",`
 border-top: 1px solid #0000;
 `)]),r("tabs-tab-pad",`
 border-top: 1px solid var(--n-tab-border-color);
 `),r("tabs-pad",`
 border-top: 1px solid var(--n-tab-border-color);
 `)])])]),r("tabs-scroll-button",[d("start",`
 padding-left: 10px;
 padding-right: 6px;
 `),d("end",`
 padding-right: 10px;
 padding-left: 6px;
 `),d("up",`
 padding-bottom: 10px;
 `),d("down",`
 padding-top: 10px;
 `)])]),rt=se({name:"TabsButton",props:{type:{type:String,default:"next"},mergedClsPrefix:{type:String,required:!0},vertical:Boolean,disabled:Boolean,rtl:Boolean,theme:Object,themeOverrides:Object,onClick:Function},setup(e){return{handleClick:()=>{var i;e.disabled||(i=e.onClick)==null||i.call(e,e.type)}}},render(){const{mergedClsPrefix:e,disabled:n,type:i,vertical:u,rtl:f,theme:R,themeOverrides:p,handleClick:b}=this,C=i==="next",g=u?C:f?!C:C;return l(),k(pe,{text:!0,disabled:n,size:"small",theme:R,themeOverrides:p,onClick:b,class:E([`${e}-tabs-scroll-button`,!u&&i==="prev"&&`${e}-tabs-scroll-button--start`,!u&&i==="next"&&`${e}-tabs-scroll-button--end`,u&&i==="prev"&&`${e}-tabs-scroll-button--up`,u&&i==="next"&&`${e}-tabs-scroll-button--down`])},{icon:()=>(l(),k(st,{clsPrefix:e,style:ie(u?{transform:"rotate(90deg)"}:void 0)},{default:()=>g?(l(),k(ka,{key:1})):(l(),k(Ja,{key:2}))},1032,["clsPrefix","style"]))},1032,["disabled","theme","themeOverrides","onClick","class"])}});const Ie=Ka,nr={...dt.props,value:[String,Number],defaultValue:[String,Number],trigger:{type:String,default:"click"},type:{type:String,default:"bar"},closable:Boolean,justifyContent:String,size:String,placement:{type:String,default:"top"},tabStyle:[String,Object],tabClass:String,addTabStyle:[String,Object],addTabClass:String,barWidth:Number,paneClass:String,paneStyle:[String,Object],paneWrapperClass:String,paneWrapperStyle:[String,Object],addable:[Boolean,Object],tabsPadding:{type:Number,default:0},animated:Boolean,onBeforeLeave:Function,onAdd:Function,"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],onClose:[Function,Array],labelSize:String,activeName:[String,Number],onActiveNameChange:[Function,Array],showScrollButton:Boolean,centerActiveTab:Boolean};var or=se({name:"Tabs",props:nr,slots:Object,setup(e,{slots:n}){var Ue,Xe;const{mergedClsPrefixRef:i,inlineThemeDisabled:u,mergedComponentPropsRef:f,mergedRtlRef:R}=da(e),p=ca("Tabs",R,i),b=le(()=>{const{placement:t}=e;return t==="start"?p!=null&&p.value?"right":"left":t==="end"?p!=null&&p.value?"left":"right":t}),C=dt("Tabs","-tabs",rr,er,e,i),g=O(null),I=O(null),s=O(null),z=O(null),_=O(null),w=O(null),T=O(null),M=O(!0),H=O(!0),Q=Je(e,["labelSize","size"]),re=le(()=>{var a,o;if(Q.value)return Q.value;const t=(o=(a=f==null?void 0:f.value)==null?void 0:a.Tabs)==null?void 0:o.size;return t||"medium"}),J=Je(e,["activeName","value"]),q=O(J.value??e.defaultValue??(n.default?(Xe=(Ue=$e(n.default())[0])==null?void 0:Ue.props)==null?void 0:Xe.name:null)),h=Pa(J,q),m={id:0},P=le(()=>{if(!(!e.justifyContent||e.type==="card"))return{display:"flex",justifyContent:e.justifyContent}});ve(h,()=>{m.id=0,F(),ge(()=>{Pe()})});function j(){var a;const{value:t}=h;return t===null?null:(a=g.value)==null?void 0:a.querySelector(`[data-name="${t}"]`)}function de(t){if(e.type==="card")return;const{value:a}=s;if(!a)return;const o=a.style.opacity==="0";if(t){const c=`${i.value}-tabs-bar--disabled`,{barWidth:L}=e,V=b.value;if(t.dataset.disabled==="true"?a.classList.add(c):a.classList.remove(c),["top","bottom"].includes(V)){if(A(["top","maxHeight","height"]),typeof L=="number"&&t.offsetWidth>=L){const W=Math.floor((t.offsetWidth-L)/2)+t.offsetLeft;a.style.left=`${W}px`,a.style.maxWidth=`${L}px`}else a.style.left=`${t.offsetLeft}px`,a.style.maxWidth=`${t.offsetWidth}px`;a.style.width="8192px",o&&(a.style.transition="none"),a.offsetWidth,o&&(a.style.transition="",a.style.opacity="1")}else{if(A(["left","maxWidth","width"]),typeof L=="number"&&t.offsetHeight>=L){const W=Math.floor((t.offsetHeight-L)/2)+t.offsetTop;a.style.top=`${W}px`,a.style.maxHeight=`${L}px`}else a.style.top=`${t.offsetTop}px`,a.style.maxHeight=`${t.offsetHeight}px`;a.style.height="8192px",o&&(a.style.transition="none"),a.offsetHeight,o&&(a.style.transition="",a.style.opacity="1")}}}function ee(){if(e.type==="card")return;const{value:t}=s;t&&(t.style.opacity="0")}function A(t){const{value:a}=s;if(a)for(const o of t)a.style[o]=""}function F(){if(e.type==="card")return;const t=j();t?de(t):ee()}function ce(t,a,o,c){const L=t.getBoundingClientRect(),V=a.getBoundingClientRect(),W=o?"left":"top",N=o?"right":"bottom";let X=0;c?X=(V[W]+V[N])/2-(L[W]+L[N])/2:V[W]<L[W]?X=V[W]-L[W]:V[N]>L[N]&&(X=V[N]-L[N]),X!==0&&t.scrollBy({[W]:X,behavior:"smooth"})}function Pe(){var o;const t=["top","bottom"].includes(b.value),a=j();if(a)if(t){const c=(o=w.value)==null?void 0:o.$el;if(!c)return;ce(c,a,t,e.centerActiveTab)}else{const{value:c}=T;if(!c)return;ce(c,a,t,e.centerActiveTab)}}const he=O(null);let ke=0,ne=null;function ft(t){const a=he.value;if(a){ke=t.getBoundingClientRect().height;const o=`${ke}px`,c=()=>{a.style.height=o,a.style.maxHeight=o};ne?(c(),ne(),ne=null):ne=c}}function ut(t){const a=he.value;if(a){const o=t.getBoundingClientRect().height,c=()=>{document.body.offsetHeight,a.style.maxHeight=`${o}px`,a.style.height=`${Math.max(ke,o)}px`};ne?(ne(),ne=null,c()):ne=c}}function pt(){const t=he.value;if(t){t.style.maxHeight="",t.style.height="";const{paneWrapperStyle:a}=e;if(typeof a=="string")t.style.cssText=a;else if(a){const{maxHeight:o,height:c}=a;o!==void 0&&(t.style.maxHeight=o),c!==void 0&&(t.style.height=c)}}}const He={value:[]},je=O("next");function vt(t){const a=h.value;let o="next";for(const c of He.value){if(c===a)break;if(c===t){o="prev";break}}je.value=o,gt(t)}function gt(t){const{onActiveNameChange:a,onUpdateValue:o,"onUpdate:value":c}=e;a&&Se(a,t),o&&Se(o,t),c&&Se(c,t),q.value=t}function ht(t){const{onClose:a}=e;a&&Se(a,t)}function mt(t){if(["top","bottom"].includes(b.value)){const{value:a}=w;if(!a)return;const o=a.$el;if(!o)return;const c=o.offsetWidth,L=!!(p!=null&&p.value),V=t==="next"?c:-c;o.scrollBy({left:L?-V:V,behavior:"smooth"})}else{const{value:a}=T;if(!a)return;const o=a.offsetHeight,c=t==="next"?a.scrollTop+o:a.scrollTop-o;a.scrollTo({top:c,left:0,behavior:"smooth"})}}let Re=!0;function ze(){const{value:t}=s;if(!t)return;Re&&(Re=!1);const a="transition-disabled";t.classList.add(a),F(),t.classList.remove(a)}const be=O(null);function me({transitionDisabled:t}){const a=g.value;if(!a)return;t&&a.classList.add("transition-disabled");const o=j();o&&be.value&&(be.value.style.width=`${o.offsetWidth}px`,be.value.style.height=`${o.offsetHeight}px`,be.value.style.transform=`translate(${o.offsetLeft}px, ${o.offsetTop}px)`,t&&be.value.offsetWidth),t&&a.classList.remove("transition-disabled")}ve([h],()=>{e.type==="segment"&&ge(()=>{me({transitionDisabled:!1})})}),ct(()=>{e.type==="segment"&&me({transitionDisabled:!0})});let De=0;function xt(t){var o;if(t.contentRect.width===0&&t.contentRect.height===0||De===t.contentRect.width)return;De=t.contentRect.width;const{type:a}=e;(a==="line"||a==="bar")&&(Re||(o=e.justifyContent)!=null&&o.startsWith("space"))&&ze(),a!=="segment"&&xe(Fe())}const yt=Ie(xt,64);function Ge(){const{type:t}=e;t==="line"||t==="bar"?ze():t==="segment"&&me({transitionDisabled:!0})}ve([()=>e.justifyContent,()=>e.size],()=>{ge(()=>{(e.type==="line"||e.type==="bar")&&ze()})}),ve([b,()=>p==null?void 0:p.value],()=>{ge(()=>{Ge(),xe(Fe(),{instantly:!0})})}),ve(()=>e.type,()=>{ge(()=>{const t=I.value;t&&(t.classList.add("transition-disabled"),Ge(),t.offsetWidth,t.classList.remove("transition-disabled"))})});const fe=O(!1);function Ct(t){var N;const{target:a,contentRect:{width:o,height:c}}=t,L=a.parentElement.parentElement.offsetWidth,V=a.parentElement.parentElement.offsetHeight,W=b.value;if(!fe.value)W==="top"||W==="bottom"?L<o&&(fe.value=!0):V<c&&(fe.value=!0);else{const{value:X}=_;if(!X)return;W==="top"||W==="bottom"?L-o>X.$el.offsetWidth&&(fe.value=!1):V-c>X.$el.offsetHeight&&(fe.value=!1)}xe(((N=w.value)==null?void 0:N.$el)||null)}const St=Ie(Ct,64);function wt(){const{onAdd:t}=e;t&&t()}const Le=O(!1);function Fe(){var a;const t=b.value;return(t==="top"||t==="bottom"?(a=w.value)==null?void 0:a.$el:T.value)||null}function xe(t,a={instantly:!1}){if(!t)return;const o=a.instantly?z.value:null;o&&o.classList.add("transition-disabled");const c=1,L=b.value;if(L==="top"||L==="bottom"){const{scrollLeft:V,scrollWidth:W,offsetWidth:N}=t,X=Math.abs(V);M.value=X<=c,H.value=X+N>=W-c,Le.value=N<W-c}else{const{scrollTop:V,scrollHeight:W,offsetHeight:N}=t;M.value=V<=c,H.value=V+N>=W-c,Le.value=N<W-c}o&&(o.offsetWidth,o.classList.remove("transition-disabled"))}const Tt=Ie(t=>{xe(t.target)},64);ga(Oe,{triggerRef:ae(e,"trigger"),tabStyleRef:ae(e,"tabStyle"),tabClassRef:ae(e,"tabClass"),addTabStyleRef:ae(e,"addTabStyle"),addTabClassRef:ae(e,"addTabClass"),paneClassRef:ae(e,"paneClass"),paneStyleRef:ae(e,"paneStyle"),mergedClsPrefixRef:i,typeRef:ae(e,"type"),closableRef:ae(e,"closable"),valueRef:h,tabChangeIdRef:m,onBeforeLeaveRef:ae(e,"onBeforeLeave"),activateTab:vt,handleClose:ht,handleAdd:wt}),La(()=>{F(),Pe()}),ba(()=>{const{value:t}=z;if(!t)return;const{value:a}=i,o=`${a}-tabs-nav-scroll-wrapper--shadow-start`,c=`${a}-tabs-nav-scroll-wrapper--shadow-end`;M.value?t.classList.remove(o):t.classList.add(o),H.value?t.classList.remove(c):t.classList.add(c)});const _t={syncBarPosition:()=>{F()},scrollToCurrentTab:()=>{Pe()}},Pt=()=>{me({transitionDisabled:!0})},Ne=le(()=>{const{value:t}=re,{type:a}=e,o=`${t}${{card:"Card",bar:"Bar",line:"Line",segment:"Segment"}[a]}`,{self:{barColor:c,closeIconColor:L,closeIconColorHover:V,closeIconColorPressed:W,tabColor:N,tabBorderColor:X,paneTextColor:kt,tabFontWeight:Rt,tabBorderRadius:zt,tabFontWeightActive:Lt,colorSegment:$t,fontWeightStrong:Bt,tabColorSegment:At,closeSize:Wt,closeIconSize:Et,closeColorHover:It,closeColorPressed:Vt,closeBorderRadius:Mt,[te("panePadding",t)]:ye,[te("tabPadding",o)]:Ot,[te("tabPaddingVertical",o)]:Ht,[te("tabGap",o)]:jt,[te("tabGap",`${o}Vertical`)]:Dt,[te("tabTextColor",a)]:Gt,[te("tabTextColorActive",a)]:Ft,[te("tabTextColorHover",a)]:Nt,[te("tabTextColorDisabled",a)]:Ut,[te("tabFontSize",t)]:Xt},common:{cubicBezierEaseInOut:Kt}}=C.value;return{"--n-bezier":Kt,"--n-color-segment":$t,"--n-bar-color":c,"--n-tab-font-size":Xt,"--n-tab-text-color":Gt,"--n-tab-text-color-active":Ft,"--n-tab-text-color-disabled":Ut,"--n-tab-text-color-hover":Nt,"--n-pane-text-color":kt,"--n-tab-border-color":X,"--n-tab-border-radius":zt,"--n-close-size":Wt,"--n-close-icon-size":Et,"--n-close-color-hover":It,"--n-close-color-pressed":Vt,"--n-close-border-radius":Mt,"--n-close-icon-color":L,"--n-close-icon-color-hover":V,"--n-close-icon-color-pressed":W,"--n-tab-color":N,"--n-tab-font-weight":Rt,"--n-tab-font-weight-active":Lt,"--n-tab-padding":Ot,"--n-tab-padding-vertical":Ht,"--n-tab-gap":jt,"--n-tab-gap-vertical":Dt,"--n-pane-padding-left":Ce(ye,"left"),"--n-pane-padding-right":Ce(ye,"right"),"--n-pane-padding-top":Ce(ye,"top"),"--n-pane-padding-bottom":Ce(ye,"bottom"),"--n-font-weight-strong":Bt,"--n-tab-color-segment":At}}),ue=u?fa("tabs",le(()=>`${re.value[0]}${e.type[0]}`),Ne,e):void 0;return{mergedClsPrefix:i,mergedValue:h,renderedNames:new Set,segmentCapsuleElRef:be,tabsPaneWrapperRef:he,tabsElRef:g,selfElRef:I,barElRef:s,addTabInstRef:_,xScrollInstRef:w,scrollWrapperElRef:z,addTabFixed:fe,tabWrapperStyle:P,handleNavResize:yt,mergedSize:re,handleScroll:Tt,handleTabsResize:St,cssVars:u?void 0:Ne,themeClass:ue==null?void 0:ue.themeClass,animationDirection:je,renderNameListRef:He,yScrollElRef:T,handleSegmentResize:Pt,onAnimationBeforeLeave:ft,onAnimationEnter:ut,onAnimationAfterEnter:pt,onRender:ue==null?void 0:ue.onRender,startReachedRef:M,endReachedRef:H,isOverflow:Le,handleButtonClick:mt,mergedTheme:C,rtlEnabled:p,mergedPlacement:b,..._t}},render(){const{mergedClsPrefix:e,type:n,mergedPlacement:i,addTabFixed:u,addable:f,mergedSize:R,renderNameListRef:p,onRender:b,paneWrapperClass:C,paneWrapperStyle:g,startReachedRef:I,endReachedRef:s,isOverflow:z,showScrollButton:_,handleButtonClick:w,mergedTheme:T,rtlEnabled:M,$slots:{default:H,prefix:Q,suffix:re}}=this;b==null||b();const J=H?$e(H()).filter(A=>A.type.__TAB_PANE__===!0):[],q=H?$e(H()).filter(A=>A.type.__TAB__===!0):[],h=!q.length,m=n==="card",P=n==="segment",j=!m&&!P&&this.justifyContent;p.value=[];const de=()=>{const A=(l(),y("div",{style:ie(this.tabWrapperStyle),class:E(`${e}-tabs-wrapper`)},[j?S(()=>null):(l(),y("div",{key:1,class:E(`${e}-tabs-scroll-padding`),style:ie(i==="top"||i==="bottom"?{width:`${this.tabsPadding}px`}:{height:`${this.tabsPadding}px`})},null,6)),h?(l(),y(Y,{key:2},[S(()=>J.map((F,ce)=>(p.value.push(F.props.name),Ve((l(),k(Me,_e(F.props,{internalCreatedByPane:!0,internalLeftPadded:ce!==0&&(!j||j==="center"||j==="start"||j==="end")}),Ke(F.children?{default:F.children.tab}:void 0),1040,["internalLeftPadded"]))))))],64)):(l(),y(Y,{key:3},[S(()=>q.map((F,ce)=>(p.value.push(F.props.name),Ve(ce!==0&&!j?it(F):F))))],64)),!u&&f&&m?(l(),y(Y,{key:4},[S(()=>ot(f,(h?J.length:q.length)!==0))],64)):S(()=>null),j?S(()=>null):(l(),y("div",{key:7,class:E(`${e}-tabs-scroll-padding`),style:ie({width:`${this.tabsPadding}px`})},null,6)),m?S(()=>null):(l(),y("div",{key:9,ref:"barElRef",class:E(`${e}-tabs-bar`)},null,2))],6));return l(),y("div",{ref:"tabsElRef",class:E(`${e}-tabs-nav-scroll-content`)},[m&&f?(l(),k(Be,{key:0,onResize:this.handleTabsResize},{default:()=>A},1032,["onResize"])):(l(),y(Y,{key:1},[S(()=>A)],64)),m?(l(),y("div",{key:2,class:E(`${e}-tabs-pad`)},null,2)):S(()=>null)],2)},ee=P?"top":i;return l(),y("div",{ref:"selfElRef",class:E([`${e}-tabs`,this.themeClass,`${e}-tabs--${n}-type`,`${e}-tabs--${R}-size`,j&&`${e}-tabs--flex`,`${e}-tabs--${ee}`,M&&`${e}-tabs--rtl`]),style:ie(this.cssVars)},[Z("div",{class:E([`${e}-tabs-nav--${n}-type`,`${e}-tabs-nav--${ee}`,`${e}-tabs-nav`])},[S(()=>Ye(Q,A=>A&&(l(),y("div",{class:E(`${e}-tabs-nav__prefix`)},[S(()=>A)],2)))),P?(l(),k(Be,{key:0,onResize:this.handleSegmentResize},{default:()=>(l(),y("div",{class:E(`${e}-tabs-rail`),ref:"tabsElRef"},[Z("div",{class:E(`${e}-tabs-capsule`),ref:"segmentCapsuleElRef"},[Z("div",{class:E(`${e}-tabs-wrapper`)},[Z("div",{class:E(`${e}-tabs-tab`)},null,2)],2)],2),h?(l(),y(Y,{key:0},[S(()=>J.map((A,F)=>(p.value.push(A.props.name),l(),k(Me,_e(A.props,{internalCreatedByPane:!0,internalLeftPadded:F!==0}),Ke(A.children?{default:A.children.tab}:void 0),1040,["internalLeftPadded"]))))],64)):(l(),y(Y,{key:1},[S(()=>q.map((A,F)=>(p.value.push(A.props.name),F===0?A:it(A))))],64))],2))},1032,["onResize"])):(l(),y(Y,{key:1},[S(()=>_&&z&&(l(),k(rt,{mergedClsPrefix:e,type:"prev",vertical:ee==="left"||ee==="right",disabled:I,rtl:!!M,theme:T.peers.Button,themeOverrides:T.peerOverrides.Button,onClick:w},null,8,["mergedClsPrefix","vertical","disabled","rtl","theme","themeOverrides","onClick"]))),(l(),k(Be,{onResize:this.handleNavResize},{default:()=>(l(),y("div",{class:E(`${e}-tabs-nav-scroll-wrapper`),ref:"scrollWrapperElRef"},[["top","bottom"].includes(ee)?(l(),k(qa,{key:0,ref:"xScrollInstRef",onScroll:this.handleScroll},{default:de},1032,["onScroll"])):(l(),y("div",{key:1,class:E(`${e}-tabs-nav-y-scroll`),onScroll:this.handleScroll,ref:"yScrollElRef"},[S(()=>de())],42,["onScroll"]))],2))},1032,["onResize"])),S(()=>_&&z&&(l(),k(rt,{mergedClsPrefix:e,type:"next",vertical:ee==="left"||ee==="right",disabled:s,rtl:!!M,theme:T.peers.Button,themeOverrides:T.peerOverrides.Button,onClick:w},null,8,["mergedClsPrefix","vertical","disabled","rtl","theme","themeOverrides","onClick"])))],64)),u&&f&&m?(l(),y(Y,{key:2},[S(()=>ot(f,!0))],64)):S(()=>null),S(()=>Ye(re,A=>A&&(l(),y("div",{class:E(`${e}-tabs-nav__suffix`)},[S(()=>A)],2))))],2),S(()=>h&&(this.animated&&(ee==="top"||ee==="bottom")?(l(),y("div",{key:1,ref:"tabsPaneWrapperRef",style:ie(g),class:E([`${e}-tabs-pane-wrapper`,C])},[S(()=>nt(J,this.mergedValue,this.renderedNames,this.onAnimationBeforeLeave,this.onAnimationEnter,this.onAnimationAfterEnter,this.animationDirection))],6)):nt(J,this.mergedValue,this.renderedNames)))],6)}});function nt(e,n,i,u,f,R,p){const b=[];return e.forEach(C=>{const{name:g,displayDirective:I,"display-directive":s}=C.props,z=w=>I===w||s===w,_=n===g;if(C.key!==void 0&&(C.key=g),_||z("show")||z("show:lazy")&&i.has(g)){i.has(g)||i.add(g);const w=!z("if");b.push(w?we(C,[[pa,_]]):C)}}),p?(l(),k(va,{name:`${p}-transition`,onBeforeLeave:u,onEnter:f,onAfterEnter:R},{default:()=>b},1032,["name","onBeforeLeave","onEnter","onAfterEnter"])):b}function ot(e,n){return l(),k(Me,{ref:"addTabInstRef",key:"__addable",name:"__addable",internalCreatedByPane:!0,internalAddable:!0,internalLeftPadded:n,disabled:typeof e=="object"&&e.disabled},null,8,["internalLeftPadded","disabled"])}function it(e){const n=ua(e);return n.props?n.props.internalLeftPadded=!0:n.props={internalLeftPadded:!0},n}function Ve(e){return Array.isArray(e.dynamicProps)?e.dynamicProps.includes("internalLeftPadded")||e.dynamicProps.push("internalLeftPadded"):e.dynamicProps=["internalLeftPadded"],e}const ir={class:"mono"},lr={class:"mono"},sr={class:"mono"},dr=se({__name:"index",setup(e){const n=ma(),i=xa(),u=wa(),{start:f,stop:R,restart:p,isRunning:b,isTransitioning:C}=Ra(),g=le(()=>n.params.id),I=O(!1),s=O(null),z=O("kv"),_=O(!1),w=ha({kv:"",nginx:""});async function T(){I.value=!0;try{const m=await Ze.get(g.value);s.value=m.data}catch(m){u.error(m.message||"加载失败"),h()}finally{I.value=!1}}async function M(){_.value=!0;try{const m=await Ze.getConfig(g.value);w.kv=m.data.kv||"",w.nginx=m.data.nginx||""}catch(m){u.error(m.message||"加载配置失败")}finally{_.value=!1}}function H(m){(m==="kv"||m==="nginx")&&!w.kv&&!w.nginx&&M()}async function Q(){await f(g.value)&&(await T(),await M())}async function re(){await R(g.value)&&await T()}async function J(){await p(g.value)&&(await T(),await M())}function q(){i.push(`/${g.value}/edit`)}function h(){i.push("/list")}return ct(async()=>{await T(),await M()}),(m,P)=>{const j=ya("permission");return l(),y("div",null,[B(v(Ae),{vertical:"",size:"large"},{default:x(()=>[B(v(qe),null,{header:x(()=>[B(v(Ae),{align:"center"},{default:x(()=>[P[1]||(P[1]=Z("span",null,"代理详情",-1)),s.value?(l(),k(v(Ca),{key:0,type:v(Ba)[s.value.status]||"default",size:"small"},{default:x(()=>[U(G(s.value.status),1)]),_:1},8,["type"])):oe("",!0)]),_:1})]),"header-extra":x(()=>[B(v(Ae),null,{default:x(()=>[s.value&&!v(b)(s.value.status)?we((l(),k(v(pe),{key:0,type:"primary",disabled:v(C)(s.value.status),onClick:Q},{default:x(()=>[...P[2]||(P[2]=[U(" 启动 ",-1)])]),_:1},8,["disabled"])),[[j,"ssl_proxy:proxy:start"]]):oe("",!0),s.value&&v(b)(s.value.status)?we((l(),k(v(pe),{key:1,type:"warning",onClick:re},{default:x(()=>[...P[3]||(P[3]=[U(" 停止 ",-1)])]),_:1})),[[j,"ssl_proxy:proxy:stop"]]):oe("",!0),s.value&&v(b)(s.value.status)?(l(),k(v(pe),{key:2,onClick:J},{default:x(()=>[...P[4]||(P[4]=[U(" 重启 ",-1)])]),_:1})):oe("",!0),we((l(),k(v(pe),{disabled:s.value?v(b)(s.value.status):!0,onClick:q},{default:x(()=>[...P[5]||(P[5]=[U(" 编辑 ",-1)])]),_:1},8,["disabled"])),[[j,"ssl_proxy:proxy:update"]]),B(v(pe),{onClick:h},{default:x(()=>[...P[6]||(P[6]=[U("返回列表",-1)])]),_:1})]),_:1})]),default:x(()=>[B(v(Ta),{show:I.value},{default:x(()=>[s.value?(l(),k(v(Sa),{key:0,bordered:"",column:2,"label-placement":"left",size:"small"},{default:x(()=>[B(v(K),{label:"service_id"},{default:x(()=>[Z("span",ir,G(s.value.service_id),1)]),_:1}),B(v(K),{label:"服务名称"},{default:x(()=>[U(G(s.value.service_name),1)]),_:1}),B(v(K),{label:"模式"},{default:x(()=>[U(G(s.value.mode==="client"?"客户端（nchttp）":"服务端（nshttp）"),1)]),_:1}),B(v(K),{label:"协议"},{default:x(()=>[U(G(v($a)[s.value.protocol]||s.value.protocol),1)]),_:1}),B(v(K),{label:"监听地址"},{default:x(()=>[Z("span",lr,G(s.value.listen_ip)+":"+G(s.value.listen_port),1)]),_:1}),B(v(K),{label:"后端地址"},{default:x(()=>[Z("span",sr,G(s.value.backend_scheme)+"://"+G(s.value.backend_ip)+":"+G(s.value.backend_port),1)]),_:1}),B(v(K),{label:"SSL Server Name"},{default:x(()=>[U(G(s.value.ssl_server_name||"—"),1)]),_:1}),B(v(K),{label:"PID"},{default:x(()=>[U(G(s.value.pid>0?s.value.pid:"—"),1)]),_:1}),B(v(K),{label:"服务器证书"},{default:x(()=>[U(G(s.value.cert_id||"—"),1)]),_:1}),B(v(K),{label:"CA 证书"},{default:x(()=>[U(G(s.value.ca_cert_id||"—"),1)]),_:1}),B(v(K),{label:"客户端证书"},{default:x(()=>[U(G(s.value.client_cert_id||"—"),1)]),_:1}),B(v(K),{label:"客户端私钥引用"},{default:x(()=>[U(G(s.value.client_key_ref||"—"),1)]),_:1}),B(v(K),{label:"创建时间"},{default:x(()=>[U(G(s.value.created_at),1)]),_:1}),B(v(K),{label:"更新时间"},{default:x(()=>[U(G(s.value.updated_at),1)]),_:1})]),_:1})):oe("",!0)]),_:1},8,["show"])]),_:1}),B(v(qe),null,{default:x(()=>[B(v(or),{value:z.value,"onUpdate:value":[P[0]||(P[0]=de=>z.value=de),H],type:"line"},{default:x(()=>[B(v(Ee),{name:"kv",tab:"KV 配置"},{default:x(()=>[z.value==="kv"?(l(),k(et,{key:0,title:"KV 配置（用户视图）",content:w.kv,loading:_.value},null,8,["content","loading"])):oe("",!0)]),_:1}),B(v(Ee),{name:"nginx",tab:"Nginx 配置"},{default:x(()=>[z.value==="nginx"?(l(),k(et,{key:0,title:"Nginx 配置（渲染结果）",content:w.nginx,loading:_.value},null,8,["content","loading"])):oe("",!0)]),_:1}),B(v(Ee),{name:"logs",tab:"日志"},{default:x(()=>[z.value==="logs"?(l(),k(Aa,{key:0,"service-id":g.value},null,8,["service-id"])):oe("",!0)]),_:1})]),_:1},8,["value"])]),_:1})]),_:1})])}}}),Cr=Wa(dr,[["__scopeId","data-v-83212414"]]);export{Cr as default};
