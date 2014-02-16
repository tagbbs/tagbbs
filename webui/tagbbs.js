var TagBBS = angular.module("TagBBS", ["ngRoute", "ngProgressLite"]);

TagBBS.config(function($routeProvider, $locationProvider) {
    $routeProvider
    .when("/login", {
        templateUrl: "login.html",
        controller: "Login"
    })
    .when("/logout", {
        templateUrl: "logout.html",
        controller: "Logout"
    })
    .when("/register", {
        templateUrl: "register.html",
        controller: "Register"
    })
    .when("/list/:query?", {
        templateUrl: "list.html",
        controller: "List"
    })
    .when("/show/:key?", {
        templateUrl: "show.html",
        controller: "Show"
    })
    .when("/put/:key?", {
        templateUrl: "put.html",
        controller: "Put"
    })
    .otherwise({redirectTo: "/login"});

    $locationProvider.html5Mode(false);
})
.controller("MainCtrl", function($scope, $location, bbs) {
    $scope.user = "";
    $scope.setUser = function(user) {
        $scope.user = user;
    };
    if (!bbs.session()) {
        if ($location.path() != "/login" && $location.path() != "/logout")
            localStorage.returnPath = $location.path();
        else
            localStorage.returnPath = "";
        $location.path("/login");
    }
    bbs.version().success(function(d) {
        $scope.name = d.result.name;
        $scope.version = d.result.version;
    });
})
.controller("Login", function($scope, $location, bbs) {
    $scope.user = "";
    $scope.pass = "";
    $scope.message = "";
    var redirect = function(d) {
        if (d.result) {
            if (localStorage.returnPath) {
                $location.path(localStorage.returnPath);
                localStorage.returnPath = "";
            } else {
                $location.path("/list");
            }
        }
    }
    $scope.submit = function() {
        bbs.login($scope.user, $scope.pass).success(function(d) {
            if (d.result) {
                localStorage.sid = d.result;
                $scope.setUser($scope.user);
            } else {
                $scope.message = "Login failed: " + d.error;
            }
        }).success(redirect);
    };

    if (localStorage.sid) {
        $scope.message = "Existing session detected, checking...";
        bbs.session(localStorage.sid);
        bbs.who().success(function(d) {
            if (d.error) {
                $scope.message = "Existing session not valid: " + d.error;
            }
        }).success(function(d) {
            $scope.setUser(d.result);
            redirect(d);
        });
    }
})
.controller("Logout", function($scope, $location, bbs) {
    $scope.error = "logging out...";
    bbs.logout().success(function(d) {
        if (d.error) {
            $scope.error = d.error;
        } else {
            localStorage.sid = "";
            $scope.setUser("");
            $location.path("/login");
        }
    });
})
.controller("Register", function($scope) {

})
.controller("List", function($scope, $routeParams, $location, bbs) {
    $scope.query = $routeParams.query || "";
    $scope.posts = [];
    $scope.newQuery = function() {
        $location.path("/list/" + $scope.query);
    };
    bbs.list($scope.query).success(function(d) {
        $scope.posts = d.result;
    });
})
.controller("Show", function($scope, $routeParams, bbs) {
    $scope.key = $routeParams.key;
    $scope.error = "";
    $scope.post = {};
    $scope.loading = true;
    $scope.show_raw = function() {
        window.open('data:text/plain;charset=utf-8,' + encodeURIComponent($scope.post.content));
    };
    $scope.$watch("key", function(key) {
        bbs.get(key).success(function(d) {
            $scope.error = d.error;
            $scope.post = d.result || {};
            $scope.loading = false;
        });
    });
})
.controller("Put", function($scope, $routeParams, $location, bbs) {
    $scope.error = "";
    $scope.key = $routeParams.key || "";
    $scope.rev = 0;
    if ($scope.key) {
        $scope.content = "Loading...";
        bbs.get($scope.key).success(function(d) {
            $scope.error = d.error;
            if (d.result) {
                $scope.rev = d.result.rev;
                $scope.content = d.result.content;
            }
        })
    } else {
        $scope.content =
            "---\n" +
            "title: \"我要饭全站\"\n" +
            "authors: [" + $scope.user + "]\n" +
            "tags: [test, 1481]\n" +
            "---\n\n" +
            "我要饭全站啦！大家快来报名！\n";
    }

    $scope.submit = function() {
        bbs.put($scope.key, $scope.rev+1, $scope.content).success(function(d) {
            if (d.error) {
                $scope.error = d.error;
            } else if (d.result) {
                $location.path("/show/" + d.result);
            }
        })
    };
})
.directive("post", function ($sce) {
    var markdown = new Showdown.converter();
    var sep = function(source) {
        source = source || "";
        var trimmed = source.trimLeft();
        if (trimmed.substring(0, 4) == "---\n") {
            var headerEnd = trimmed.indexOf("\n---\n");
            if (headerEnd > 0) {
                var header = trimmed.substring(0, headerEnd);
                var body = trimmed.substring(headerEnd + 5);
                try {
                    return {header: jsyaml.safeLoad(header), body: body}
                } catch (e) {
                    console.log(e);
                }
            }
        }
        return {header: null, body: source}
    }
    return {
        require: "ngModel",
        restrict: "E",
        templateUrl: "post.html",
        link: function (scope, element, attrs, ngModel) {
            ngModel.$render = function() {
                var post = ngModel.$viewValue;
                scope.rev = post.rev;
                scope.timestamp = post.timestamp;
                scope.title = "";
                scope.tags = [];
                scope.authors = [];
                scope.body = "";
                scope.source = post.content;

                var hb = sep(post.content);
                if (hb.header) {
                    scope.title = hb.header.title;
                    scope.tags = hb.header.tags;
                    scope.authors = hb.header.authors;
                    if (hb.header.raw) scope.source = hb.body;
                    else scope.body = $sce.trustAsHtml(markdown.makeHtml(hb.body));
                }
            };
        }
    };
})
.directive("codemirror", function(isMobile) {
return {
        require: "ngModel",
        restrict: "A",
        link: function (scope, elm, attrs, ngModel) {
            // Disable codemirror for mobile. The touch interface does not seem play well with it.
            if (isMobile.any()) return;
            var codemirror = CodeMirror.fromTextArea(elm[0])
            codemirror.on("change", function(mirror) {
                var newValue = mirror.getValue();
                if (newValue !== ngModel.$viewValue) {
                    ngModel.$setViewValue(newValue);
                    scope.$apply();
                }
            });
            ngModel.$render = function() {
                codemirror.setValue(ngModel.$viewValue);
            };
       }
    };
})
.factory("bbs", function($http, serviceEndpoint) {
    var sid = "";
    var api = function(name, data) {
        data = data || {};
        data.session = sid;
        var promise = $http({
            method: 'POST',
            url: serviceEndpoint + "/" + name,
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            transformRequest: function(obj) {
                var str = [];
                for(var p in obj)
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
                return str.join("&");
            },
            data: data
        });
        return promise;
    };
    return {
        login: function(user, pass) {
            return api("login", {user: user, pass: pass}).success(function(d) {
                if (d.result) {
                    sid = d.result;
                }
                return d;
            });
        },
        logout: function() {
            return api("logout");
        },
        version: function() {
            return api("version");
        },
        who: function() {
            return api("who");
        },
        list: function(query) {
            return api("list", {query: query});
        },
        get: function(key) {
            return api("get", {key: key});
        },
        put: function(key, rev, content) {
            return api("put", {key:key, rev: rev, content: content});
        },
        session: function(_sid) {
            oldsid = sid;
            if (typeof _sid != 'undefined') {
                sid = _sid
            }
            return oldsid;
        }
    };
})
.config(function($httpProvider) {
    $httpProvider.interceptors.push(function($q, $timeout, ngProgressLite) {
        var active = 0;
        var start = function() {
            active++;
            $timeout(function() {
                if (active > 0) {
                    ngProgressLite.start();
                    ngProgressLite.inc();
                }
            }, 100);
        };
        var finish = function() {
            active--;
            $timeout(function() {
                if (active == 0) {
                    ngProgressLite.done();
                }
            }, 100);
        };
        return {
            'request': function(config) {
              start()
              return config || $q.when(config);
            },
            'response': function(response) {
              finish();
              return response || $q.when(response);
            },
           'responseError': function(rejection) {
              finish();
              return $q.reject(rejection);
            }
        };
    });
})
.value("serviceEndpoint", location.protocol + "//" + location.hostname + ":8023")
.factory("isMobile", function() {
    var isMobile = {
        Android: function() {
            return navigator.userAgent.match(/Android/i);
        },
        BlackBerry: function() {
            return navigator.userAgent.match(/BlackBerry/i);
        },
        iOS: function() {
            return navigator.userAgent.match(/iPhone|iPad|iPod/i);
        },
        Opera: function() {
            return navigator.userAgent.match(/Opera Mini/i);
        },
        Windows: function() {
            return navigator.userAgent.match(/IEMobile/i);
        },
        any: function() {
            console.log(navigator.userAgent)
            return (isMobile.Android() || isMobile.BlackBerry() || isMobile.iOS() || isMobile.Opera() || isMobile.Windows());
        }
    };
    return isMobile;
})
;
