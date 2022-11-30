var q;
var nextPage=false;
$('#quizTime').text(new Date().toLocaleTimeString([], { hour: '2-digit', minute: "2-digit", hour12: true }))

function setCount(d,n) {
  
    let ls=window.localStorage;
    if(ls.getItem('question')<8){
       ls.setItem('question',parseInt(ls.getItem('question'))+1) 
    }
    
    if(d===answer){
        ls.setItem('correct',parseInt(ls.getItem('correct'))+1)
    }
    

            Swal.fire(
                "Success",
                "your answer is submited successfully!",
                "success"
            ),
            $('button:contains("OK")').on("click", function () {
               if(parseInt(ls.getItem('question'))===8){
                window.location.href='/certificate?rid='+ls.getItem('rid')
                Swal.fire({
                    title: 'Please Wait !',
                    html: 'we are creating your certificate',// add html attribute if you want or remove
                    allowOutsideClick: false,
                    onBeforeOpen: () => {
                        Swal.showLoading()
                    },
                });
                

            }else{
                window.location.href='/quiz?id='+ls.getItem('question')+"&rid="+ls.getItem('rid')
            } 
            });
    
}





