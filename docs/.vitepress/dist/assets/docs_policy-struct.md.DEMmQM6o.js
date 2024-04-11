import{_ as s,c as i,o as a,a2 as n}from"./chunks/framework.D_Kv-6xY.js";const F=JSON.parse('{"title":"A starting point","description":"","frontmatter":{},"headers":[],"relativePath":"docs/policy-struct.md","filePath":"docs/policy-struct.md"}'),t={name:"docs/policy-struct.md"},l=n(`<h1 id="a-starting-point" tabindex="-1">A starting point <a class="header-anchor" href="#a-starting-point" aria-label="Permalink to &quot;A starting point&quot;">​</a></h1><p>create your first policy file (mypolicy.yaml) and add the following :</p><div class="language-yaml vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">yaml</span><pre class="shiki shiki-themes github-light github-dark vp-code"><code><span class="line"></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">Banner</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">|</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">  | Starting point 1 SCAN and 1 COLLECT RULE</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">Rules</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">:</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">  - </span><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">name</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">Passwords being used in URIs</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    id</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">100</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    description</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">Detecting the pattern &quot;protocol://username:password@host&quot;</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    error</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">This violation immediately blocks your code deployment</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    tags</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">URI</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    type</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">scan</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    fatal</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">true</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    enforcement</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">true</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    environment</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">all</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    confidence</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">high</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    patterns</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">:</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">      - </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">\\s*^(.*):\\/\\/([^:]*):([^@]*)@(.*)$</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">  - </span><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">name</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">Collect proxy modifications on your bootstrap</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    id</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">800</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    description</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">The following proxy modifications were collected</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    type</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">collect</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    tags</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">AWS,AZURE</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">    patterns</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">:</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">      - </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">\\b(?:http_proxy|https_proxy|ftp_proxy|socks_proxy|no_proxy|HTTP_PROXY|HTTPS_PROXY|FTP_PROXY|SOCKS_PROXY|NO_PROXY)\\s*=\\s*[&#39;&quot;]?(https?|socks[45])://(?:[^\\s&#39;&quot;]+)</span></span>
<span class="line"></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">ExitCritical</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;Critical irregularities found in your code&quot;</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">ExitWarning</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;Irregularities found in your code&quot;</span></span>
<span class="line"><span style="--shiki-light:#22863A;--shiki-dark:#85E89D;">ExitClean</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;Clean report&quot;</span></span></code></pre></div><h2 id="policy-struct-schema" tabindex="-1">Policy Struct Schema <a class="header-anchor" href="#policy-struct-schema" aria-label="Permalink to &quot;Policy Struct Schema&quot;">​</a></h2><div class="language-go vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">go</span><pre class="shiki shiki-themes github-light github-dark vp-code"><code><span class="line"><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">type</span><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;"> Rule</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;"> struct</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> {</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	ID               </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">int</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">      \`yaml:&quot;id&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Name             </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;name&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Description      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;description&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Solution         </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;solution&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Error            </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;error&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Type             </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;type&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Environment      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;environment&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Enforcement      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">bool</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">     \`yaml:&quot;enforcement&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Fatal            </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">bool</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">     \`yaml:&quot;fatal&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    </span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Tags             </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;tags,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Impact           </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;impact,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Confidence       </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;confidence,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    Api_Endpoint     </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;api_endpoint,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Request      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;api_request,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Insecure     </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">bool</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">     \`yaml:&quot;api_insecure&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Body         </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;api_body,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Auth         </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;api_auth,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Auth_Basic   </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">*</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">  \`yaml:&quot;api_auth_basic,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Auth_Token   </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">*</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">  \`yaml:&quot;api_auth_token,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Api_Trace        </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">bool</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">     \`yaml:&quot;api_trace,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    Filepattern      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;filepattern,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    Yml_Filepattern  </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;yml_filepattern,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Yml_Structure    </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;yml_structure,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    Toml_Filepattern </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;toml_filepattern,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Toml_Structure   </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;toml_structure,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    Json_Filepattern </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;json_filepattern,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Json_Structure   </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;json_structure,omitempty&quot;\`</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Rego_Filepattern      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;rego_filepattern,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Rego_Policy_File      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;rego_policy_file,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Rego_Policy_Data      </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;rego_policy_data,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Rego_Policy_Query     </span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">   \`yaml:&quot;rego_policy_query,omitempty&quot;\`</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">	Patterns              []</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">string</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> \`yaml:&quot;patterns,omitempty&quot;\`</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">}</span></span></code></pre></div><h2 id="run-it" tabindex="-1">Run it <a class="header-anchor" href="#run-it" aria-label="Permalink to &quot;Run it&quot;">​</a></h2><div class="language-sh vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">sh</span><pre class="shiki shiki-themes github-light github-dark vp-code"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> pull</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/xfhg/intercept:latest</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> run</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -v</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --rm</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -w</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-v</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">:</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">$PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-e</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> TERM=xterm-256color</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/xfhg/intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> config</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -r</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> run</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -v</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --rm</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -w</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-v</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">:</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">$PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-e</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> TERM=xterm-256color</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/xfhg/intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> config</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -a</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> mypolicy.yaml</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">docker</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> run</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -v</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --rm</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -w</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-v</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> $PWD</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">:</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">$PWD </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">-e</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> TERM=xterm-256color</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ghcr.io/xfhg/intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> intercept</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> audit</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -t</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> yourtargetfolder/</span></span></code></pre></div>`,7),h=[l];function p(k,e,r,E,y,d){return a(),i("div",null,h)}const o=s(t,[["render",p]]);export{F as __pageData,o as default};
