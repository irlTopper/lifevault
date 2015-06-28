define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        # Handle Params
        @title = params.title || "Upload Image"
        @text = params.text || "Choose image to upload"
        @subText = params.subText || ""
        @callback = params.callback
        @id = params.id
        @fileType  = params.fileType
        if not params.fileType?
            return
        if not params.id?
            return

        # Vars part
        @imageUploadURL = "v1/upload/" + @fileType
        @isUploading = ko.observable(false)

        # Setup OnShow
        app.modal.Init('imageupload', @, params)
        return

    VM::OnShow = () ->
        # Watch the iframe that does the upload for changes
        $('#uploadFrame').on 'load', =>
            res = $.parseJSON($('#uploadFrame').contents().text())
            upload = res.upload
            err = res.errors

            if err? and err.length > 0
                app.flash.Error("<strong>Error uploading file</strong> &mdash; #{err[0]}")
                return

            if @callback? then @callback upload.URL
            @isUploading(false)
            @Close()
            return
        return

    VM::Close = () ->
        app.modal.Close(@, false)
        return

    VM::Cancel = ->
        @Close(@, false)
        return

    VM::Save = ->
        $("#uploadImageForm").submit()
        return

    VM::OnSelectImage = () ->
        @isUploading(true)
        @Save()
        return

    return {
        viewModel: VM
        template: templateMarkup
    }