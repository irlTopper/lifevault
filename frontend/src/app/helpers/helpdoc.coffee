define ['knockout'], (ko) ->

    Helpdoc = () ->
        # All the helpdoc URL and title
        @helpdocData = {
            cname: {
                url: "getting-started-guide-183/setup-a-custom-domain-using-a-cname",
                title: "Setup a custom domain using a CNAME"
            }
            forwarding: {
                url: "general-settings/how-do-i-set-a-forwarding-status"
                title: "Setting up mail forwarding"
            }
        }

        # Var
        @baseUrl = "http://lifevault.helpdocs.com/"
        @iframeSuffix = "?_template=iframepreview"
        return

    Helpdoc::GetIframeUrl = (key, isIframe) ->
        url = @baseUrl + @helpdocData[key].url

        if isIframe
            url += @iframeSuffix

        return url

    Helpdoc::GetTitle = (key) ->
        return @helpdocData[key].title

    Helpdoc::ShowModal = (key) ->
        key = ko.utils.unwrapObservable(key)
        if Number(key) > 0
            $.ajax
                url: "v1/helpdocs/articles/#{key}.json"
                dataType: "json"
                method: "GET"
                success: (data) =>
                    app.modal.Show "helpdoc-content", {
                        article: data.article
                        urlIframe: "#{data.article.previewUrl}?_template=iframepreview"
                        editable: !@helpdocData[key]?
                    }
                error: (xhr) =>
                    app.error.Ajax(xhr)
        else
            unless @helpdocData[key]
                return false

            app.modal.Show "helpdoc-content", {
                article:
                    title: @helpdocData[key].title
                    publicUrl: @GetIframeUrl(key)
                urlIframe: @GetIframeUrl(key, true)
                urlHelpDoc: @GetIframeUrl(key)
                editable: false
            }

    return Helpdoc