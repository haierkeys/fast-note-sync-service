import{c as a}from"./theme-switcher-DO7hUccB.js";import{o as s,c as e,s as i,l as o}from"./zod-CR08G2-F.js";
/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const n=a("Github",[["path",{d:"M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4",key:"tonef"}],["path",{d:"M9 18c-4.51 2-5-2-7-2",key:"9comsn"}]]),r=a=>i().min(6,a("ui.auth.passwordMinLength")),t=a=>s({credentials:i().min(1,a("ui.auth.credentialsRequired")),password:r(a),remember:e.boolean().optional().default(!1)}),m=a=>s({username:i().min(3,a("ui.auth.usernameMinLength")),email:i().email(a("ui.auth.emailInvalid")),password:r(a),confirmPassword:i().min(6,a("ui.auth.passwordMinLength"))}).refine(a=>a.password===a.confirmPassword,{message:a("ui.auth.passwordMismatch"),path:["confirmPassword"]}),d=a=>s({username:i().min(3,a("ui.auth.usernameMinLength")),email:i().email(a("ui.auth.emailInvalid")),password:r(a).or(o("")),isDeleted:e.boolean().optional().default(!1)});s({oldPassword:i().optional(),password:i().optional(),confirmPassword:i().optional()});export{n as G,m as a,d as b,t as c};
