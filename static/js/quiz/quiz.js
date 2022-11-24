var q;
$('#quizTime').text(new Date().toLocaleTimeString([], { hour: '2-digit', minute: "2-digit", hour12: true }))

function setCount(d,n) {
  
    let ls=window.localStorage;
    ls.setItem('question',parseInt(ls.getItem('question'))+1)
    if(d===answer){
        ls.setItem('correct',parseInt(ls.getItem('correct'))+1)
    }
    if(parseInt(ls.getItem('question'))===8){
        window.location.href='/quiz/certificate?rid='+ls.getItem('rid')
        
    }else{
        window.location.href='/quiz?id='+ls.getItem('question')+"&rid="+ls.getItem('rid')
    }
}




