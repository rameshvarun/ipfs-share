Dropzone.options.dropzone = {
	init: function() {
		this.on("success", function(file, response) {
			window.open(response.url, '_blank')
		});
	}
}

$(function() {
	$("#pastebin-form").submit(function(e) {
		e.preventDefault();

		var file_content = $("#pastebin-content").val()
		$.ajax({
			type: "POST",
			dataType: "json",
			data: {
				content : $("#pastebin-content").val()
			},
			url: "/paste",
			success: function(data) {
				$("#pastebin-content").val("");
				window.open(data.url, '_blank')
			},
			error : function(jqXHR, status, message) {
				sweetAlert(status, message, "error");
			}
		});
		
	})
})

$("html").pasteImageReader(function(results) {
	var dataURL, filename;
 	filename = results.filename, dataURL = results.dataURL;

	$.ajax({
		type: "POST",
		dataType: "json",
		data: {
			dataurl : dataURL
		},
		url: "/image",
		success: function(data) {
			window.open(data.url, '_blank')
		},
		error : function(jqXHR, status, message) {
			sweetAlert(status, message, "error");
		}
	});
});