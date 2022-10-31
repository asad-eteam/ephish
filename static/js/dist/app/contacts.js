$(document).ready(function () {
    api.contacts.get()
    .success(function (c){
    console.log('cccccccccccc',c.data.length);
    contacts=c
    if (c.data!==null){
       $('#contacts').append('<tr rol="row"><td>'+c.data.username+'</td><td>'+c.data.campaign+'</td><td>'+c.data.message+'</td></tr>');  
    }
    })
}); 