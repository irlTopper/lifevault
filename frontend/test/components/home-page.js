define(['components/dashboard-page/dashboard-page'], function(homePage) {
  var HomePageViewModel = homePage.viewModel;

  describe('Dashboard page view model', function() {

    it('should supply a friendly message which changes when acted upon', function() {
      var instance = new HomePageViewModel();
      expect(instance.message()).toContain('Welcome to ');

      // See the message change
      instance.doSomething();
      expect(instance.message()).toContain('You invoked doSomething()');
    });

  });

});
