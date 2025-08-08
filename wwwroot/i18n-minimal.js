const i18n={
    current:localStorage.getItem('lang')||(navigator.language||navigator.userLanguage||'en').startsWith('zh')?'zh':'en',
    translations:{
        en:{
            title:'📋 Web Clipboard','expiry-notice':'⏰ Content expires after 10 minutes','text-clipboard':'📝 Text','file-clipboard':'📁 File','recent-items':'📋 Recent','text-placeholder':'Paste or type text...','text-id-placeholder':'Enter 4-char ID','file-id-placeholder':'Enter 4-char ID','save-text':'💾 Save','copy-text':'📋 Copy','load-text':'📥 Load','select-file':'📎 Select file','upload-file':'⬆️ Upload','download-file':'📥 Download',
            'enter-text':'Enter text','saved':'Saved! ID: {0}','failed':'Failed','error':'Error: {0}','no-text':'No text','copied':'Copied!','copy-failed':'Copy failed','enter-id':'Enter ID','loaded':'Loaded!','not-found':'Not found','select-file-msg':'Select file','uploaded':'Uploaded! ID: {0}','downloaded':'Downloaded!','id-copied':'ID copied!','no-recent':'No recent items'
        },
        zh:{
            title:'📋 网页剪贴板','expiry-notice':'⏰ 内容10分钟后过期','text-clipboard':'📝 文本','file-clipboard':'📁 文件','recent-items':'📋 最近','text-placeholder':'粘贴或输入文本...','text-id-placeholder':'输入4位ID','file-id-placeholder':'输入4位ID','save-text':'💾 保存','copy-text':'📋 复制','load-text':'📥 加载','select-file':'📎 选择文件','upload-file':'⬆️ 上传','download-file':'📥 下载',
            'enter-text':'请输入文本','saved':'已保存！ID：{0}','failed':'失败','error':'错误：{0}','no-text':'无文本','copied':'已复制！','copy-failed':'复制失败','enter-id':'请输入ID','loaded':'已加载！','not-found':'未找到','select-file-msg':'请选择文件','uploaded':'上传成功！ID：{0}','downloaded':'下载成功！','id-copied':'ID已复制！','no-recent':'暂无最近项目'
        }
    },
    init(){
        const enBtn=document.getElementById('langEn');
        const zhBtn=document.getElementById('langZh');
        if(enBtn&&zhBtn){
            enBtn.onclick=()=>this.setLang('en');
            zhBtn.onclick=()=>this.setLang('zh');
        }
        this.updateUI();
    },
    setLang(lang){
        this.current=lang;
        localStorage.setItem('lang',lang);
        document.documentElement.lang=lang==='zh'?'zh-CN':'en';
        this.updateUI();
    },
    updateUI(){
        document.querySelectorAll('[data-i18n]').forEach(el=>{
            const key=el.getAttribute('data-i18n');
            const text=this.translations[this.current][key];
            if(text)el.textContent=text;
        });
        document.querySelectorAll('[data-i18n-placeholder]').forEach(el=>{
            const key=el.getAttribute('data-i18n-placeholder');
            const text=this.translations[this.current][key];
            if(text)el.placeholder=text;
        });
        const enBtn=document.getElementById('langEn');
        const zhBtn=document.getElementById('langZh');
        if(enBtn&&zhBtn){
            enBtn.className='lang-btn '+(this.current==='en'?'btn-blue':'');
            zhBtn.className='lang-btn '+(this.current==='zh'?'btn-blue':'');
            enBtn.style.background=this.current==='en'?'#3b82f6':'#f3f4f6';
            enBtn.style.color=this.current==='en'?'#fff':'#374151';
            zhBtn.style.background=this.current==='zh'?'#3b82f6':'#f3f4f6';
            zhBtn.style.color=this.current==='zh'?'#fff':'#374151';
        }
    },
    t(key,...args){
        let text=this.translations[this.current][key]||key;
        args.forEach((arg,i)=>text=text.replace(`{${i}}`,arg));
        return text;
    }
};
if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',()=>i18n.init());
else i18n.init();