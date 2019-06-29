class AppearHandler {
  constructor(selector, opts = {"onappear" : function(e){}, "ondisappear" : function(e){}}) {
      this.selectors = [
          /* None at the moment */
      ];
      this.checkBinded = false;
      this.checkLock = false;
      this.defaults = {
          interval: 250,
          force_process: false
      };
      this.priorAppeared = [];
      if (typeof(selector) != "undefined") {
        this.addSelector(selector, opts)
      }
      this.startMonitorLoop();
  }

  process () {
    this.checkLock = false;


    function isVisible ( element ) {
        return !!( element.offsetWidth || element.offsetHeight || element.getClientRects().length );
    }

    function isAppeared(element) {

        if (!isVisible(element)) {
            return false;
        }

        let windowLeft = window.scrollX;
        let windowTop = window.scrollY;
        let left = element.offsetLeft;
        let top = element.offsetTop;

        let belowTopEdge = top + element.clientHeight >= windowTop;
        let aboveBottomEdge = top  <= windowTop + window.innerHeight;

        let rightAfterLeftEdge = left + element.clientWidth >= windowLeft;
        let leftBeforeRightEdge = left <= windowLeft + window.innerWidth;

        element.dataset.belowTopEdge = belowTopEdge;

        return belowTopEdge && aboveBottomEdge && rightAfterLeftEdge && leftBeforeRightEdge;
    }

    for (var index = 0, selectorsLength = this.selectors.length; index < selectorsLength; index++) {
      var appeared = this.selectors[index].filter(isAppeared);

      appeared
        .filter(element => element.dataset._appearTriggered != "true")
        .map(element => element.dispatchEvent(new Event("appear")));

      if (this.priorAppeared[index]) {
        var disappeared = this.priorAppeared[index].filter(x => !appeared.includes(x));
        disappeared.filter(element => element.dataset._appearTriggered == "true")
            .map(element => element.dispatchEvent(new Event("disappear")));
      }
      this.priorAppeared[index] = appeared;
    }
  }

  addSelector (selector, opts = {"onappear" : function(e){}, "ondisappear" : function(e){}}) {
      let elements = Array.prototype.slice.call(document.querySelectorAll(selector));

      this.priorAppeared.push(elements);
      this.selectors.push(elements);

      elements.map(element => element.addEventListener('appear', function (e) {
          element=e.target;
          element.dataset._appearTriggered = "true";
          opts.onappear(element);
      }));
      elements.map(element => element.addEventListener('disappear', function (e) {
          element=e.target;
          element.dataset._appearTriggered = "false";
          opts.ondisappear(element);
      }));

      this.checkAfterFewSeconds();
  }

  checkAfterFewSeconds() {
      if (this.checkLock) {
          return;
      }
      this.checkLock = true;

      setTimeout(this.process.bind(this), this.defaults.interval);
  }

  startMonitorLoop () {
      if (!this.checkBinded) {
        window.addEventListener('scroll', this.checkAfterFewSeconds.bind(this));
        window.addEventListener('resize', this.checkAfterFewSeconds.bind(this));
        this.checkBinded = true;
      }
    }
}
