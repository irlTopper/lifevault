define [
    'ohlife2',
    'knockout',
    'jquery',
    'jquery-ui',
    'jquery.validate',
    'knockout-projections',
    'knockout-mapping',
    'lscache',
    'redactor',
    'bootstrap',
    'modernizr',
    'app/config/components',
], (ohlife2, ko) ->

    ## Bada boom - create & start the application, assign to window for easy access to "app" from anywhere
    (window.app = new ohlife2()).init()