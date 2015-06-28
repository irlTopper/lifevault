define [], ->

    Redactor = () ->
        return

    Redactor::CommonSettings = ->
        return {
            buttonSource: true
            linebreaks: true
            plugins: ["clipboardPaste"]
            clipboardUploadUrl: 'v1/clipboard/image.json'
            clipboardUpload: true
        }

    Redactor::MergeSettings = (specificSettings) ->
        plugins = specificSettings.plugins
        commonSettings = app.redactor.CommonSettings()

        mergedSettings = $.extend( specificSettings, commonSettings )

        if ( plugins? && plugins.length != 0)
            mergedSettings.plugins.push.apply(mergedSettings.plugins, plugins)

        return mergedSettings

    return Redactor