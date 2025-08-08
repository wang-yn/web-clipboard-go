class WebClipboard{
    constructor(){
        this.recent=JSON.parse(localStorage.getItem('recent')||'[]');
        this.init();
    }
    
    init(){
        this.bindEvents();
        this.updateRecent();
    }
    
    bindEvents(){
        const d=document;
        d.getElementById('saveText').onclick=()=>this.saveText();
        d.getElementById('copyText').onclick=()=>this.copyText();
        d.getElementById('loadText').onclick=()=>this.loadText();
        d.getElementById('selectFile').onclick=()=>d.getElementById('fileInput').click();
        d.getElementById('fileInput').onchange=e=>this.handleFile(e);
        d.getElementById('uploadFile').onclick=()=>this.uploadFile();
        d.getElementById('downloadFile').onclick=()=>this.downloadFile();
        d.getElementById('textId').onkeypress=e=>e.key==='Enter'&&this.loadText();
        d.getElementById('fileId').onkeypress=e=>e.key==='Enter'&&this.downloadFile();
        
        const f=d.getElementById('selectFile');
        f.ondragover=e=>{e.preventDefault();f.style.borderColor='#3b82f6'};
        f.ondragleave=()=>f.style.borderColor='#ddd';
        f.ondrop=e=>{
            e.preventDefault();
            f.style.borderColor='#ddd';
            const files=e.dataTransfer.files;
            if(files.length){
                d.getElementById('fileInput').files=files;
                this.handleFile({target:{files}});
            }
        };
    }
    
    async saveText(){
        const content=document.getElementById('textContent').value.trim();
        if(!content)return this.showMsg(i18n.t('enter-text'),'error');
        
        try{
            const r=await fetch('/api/text',{
                method:'POST',
                headers:{'Content-Type':'application/json'},
                body:JSON.stringify({content})
            });
            if(r.ok){
                const data=await r.json();
                this.showMsg(i18n.t('saved',data.id),'success');
                this.addRecent('text',data.id,content.substring(0,50)+'...',data.expiresAt);
                document.getElementById('textId').value=data.id;
            }else throw new Error(i18n.t('failed'));
        }catch(e){
            this.showMsg(i18n.t('error',e.message),'error');
        }
    }
    
    async copyText(){
        const content=document.getElementById('textContent').value;
        if(!content)return this.showMsg(i18n.t('no-text'),'error');
        
        try{
            await navigator.clipboard.writeText(content);
            this.showMsg(i18n.t('copied'),'success');
        }catch{
            this.showMsg(i18n.t('copy-failed'),'error');
        }
    }
    
    async loadText(){
        const id=document.getElementById('textId').value.trim();
        if(!id)return this.showMsg(i18n.t('enter-id'),'error');
        
        try{
            const r=await fetch(`/api/text/${id}`);
            if(r.ok){
                const data=await r.json();
                document.getElementById('textContent').value=data.content;
                this.showMsg(i18n.t('loaded'),'success');
            }else if(r.status===404){
                this.showMsg(i18n.t('not-found'),'error');
            }else throw new Error(i18n.t('failed'));
        }catch(e){
            this.showMsg(i18n.t('error',e.message),'error');
        }
    }
    
    handleFile(e){
        const file=e.target.files[0];
        const display=document.getElementById('selectedFile');
        const btn=document.getElementById('uploadFile');
        
        if(file){
            display.textContent=`${file.name} (${(file.size/1024/1024).toFixed(2)}MB)`;
            btn.disabled=false;
        }else{
            display.textContent='';
            btn.disabled=true;
        }
    }
    
    async uploadFile(){
        const file=document.getElementById('fileInput').files[0];
        if(!file)return this.showMsg(i18n.t('select-file-msg'),'error');
        
        const fd=new FormData();
        fd.append('file',file);
        
        try{
            const r=await fetch('/api/file',{method:'POST',body:fd});
            if(r.ok){
                const data=await r.json();
                this.showMsg(i18n.t('uploaded',data.id),'success');
                this.addRecent('file',data.id,data.fileName,data.expiresAt);
                document.getElementById('fileId').value=data.id;
            }else throw new Error(i18n.t('failed'));
        }catch(e){
            this.showMsg(i18n.t('error',e.message),'error');
        }
    }
    
    async downloadFile(){
        const id=document.getElementById('fileId').value.trim();
        if(!id)return this.showMsg(i18n.t('enter-id'),'error');
        
        try{
            const r=await fetch(`/api/file/${id}`);
            if(r.ok){
                const blob=await r.blob();
                const url=URL.createObjectURL(blob);
                const a=document.createElement('a');
                const cd=r.headers.get('content-disposition');
                let filename='download';
                if(cd){
                    const match=cd.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
                    if(match)filename=match[1].replace(/['"]/g,'');
                }
                a.href=url;
                a.download=filename;
                a.click();
                URL.revokeObjectURL(url);
                this.showMsg(i18n.t('downloaded'),'success');
            }else if(r.status===404){
                this.showMsg(i18n.t('not-found'),'error');
            }else throw new Error(i18n.t('failed'));
        }catch(e){
            this.showMsg(i18n.t('error',e.message),'error');
        }
    }
    
    addRecent(type,id,desc,exp){
        const item={type,id,desc,created:new Date().toISOString(),exp};
        this.recent=[item,...this.recent.filter(i=>i.id!==id)].slice(0,5);
        localStorage.setItem('recent',JSON.stringify(this.recent));
        this.updateRecent();
    }
    
    updateRecent(){
        const c=document.getElementById('recentItems');
        const now=new Date();
        const valid=this.recent.filter(i=>new Date(i.exp)>now);
        
        if(valid.length!==this.recent.length){
            this.recent=valid;
            localStorage.setItem('recent',JSON.stringify(this.recent));
        }
        
        if(!valid.length){
            c.innerHTML=`<p style="text-align:center;color:#666">${i18n.t('no-recent')}</p>`;
            return;
        }
        
        c.innerHTML=valid.map(i=>`
            <div class="item">
                <div>
                    <span>${i.type==='text'?'📝':'📁'}</span> ${i.desc}<br>
                    <small style="color:#666">ID: ${i.id}</small>
                </div>
                <div>
                    <button class="item-btn btn-blue" onclick="app.copyId('${i.id}')">📋</button>
                    <button class="item-btn btn-green" onclick="app.loadItem('${i.type}','${i.id}')">📥</button>
                </div>
            </div>
        `).join('');
    }
    
    async copyId(id){
        try{
            await navigator.clipboard.writeText(id);
            this.showMsg(i18n.t('id-copied'),'success');
        }catch{
            this.showMsg(i18n.t('copy-failed'),'error');
        }
    }
    
    async loadItem(type,id){
        if(type==='text'){
            document.getElementById('textId').value=id;
            await this.loadText();
        }else{
            document.getElementById('fileId').value=id;
            await this.downloadFile();
        }
    }
    
    showMsg(msg,type){
        const el=document.getElementById('statusMessage');
        el.textContent=msg;
        el.className=`status ${type}`;
        el.style.display='block';
        setTimeout(()=>el.style.display='none',3000);
    }
}

let app;function initApp(){if(typeof i18n!=='undefined'&&document.readyState!=='loading'){app=new WebClipboard()}else{setTimeout(initApp,10)}}initApp()