function load() {
    api.quizId.get(1)
        .success(function (questions) {
            console.log('qqqqqqqqqqqqqqqq',questions)
        })
        .error(function () {
        })
}
$(document).ready(function () {
    load()
})
