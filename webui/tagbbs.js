var TagBBS = angular.module("TagBBS", ["ngRoute"]);

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
    if (!bbs.session()) {
        localStorage.returnPath = $location.path();
        $location.path("/login");
    }
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
        }).success(redirect);
    }
})
.controller("Logout", function($scope, $location, bbs) {
    $scope.error = "logging out...";
    bbs.logout().success(function(d) {
        if (d.error) {
            $scope.error = d.error;
        } else {
            localStorage.sid = "";
            $location.path('/login');
        }
    });
})
.controller("Register", function($scope) {

})
.controller("List", function($scope, $routeParams, bbs) {
    $scope.query = $routeParams.query || "";
    $scope.posts = [];
    $scope.$watch("query", function(q) {
        bbs.list($scope.query).success(function(d) {
            $scope.posts = d.result;
        })
    });
})
.controller("Show", function($scope, $routeParams, bbs) {
    $scope.key = $routeParams.key;
    $scope.revision = 0;
    $scope.timestamp = null;
    $scope.content = "";
    $scope.error = "";
    $scope.$watch("key", function(key) {
        bbs.get(key).success(function(d) {
            $scope.error = d.error;

            var post = d.result || {};
            $scope.revision = post.rev;
            $scope.timestamp = post.timestamp;
            $scope.content = post.content;
        });
    });
})
.controller("Put", function($scope, $routeParams, $location, bbs) {
    $scope.key = $routeParams.key || "";
    $scope.content = "";
    $scope.error = "";

    $scope.submit = function() {
        bbs.put($scope.key, $scope.content).success(function(d) {
            if (d.error) {
                $scope.error = d.error;
            } else if (d.result) {
                $location.path("/show/" + d.result);
            }
        })
    };
})
.directive("post", function ($compile, $http) {
    var markdown = new Showdown.converter();
    var convert = function(source) {
        if (!source) return "";
        var trimmed = source.trimLeft();
        if (trimmed.substring(0, 4) == "---\n") {
            var headerEnd = trimmed.indexOf("\n---\n");
            if (headerEnd > 0) {
                var header = trimmed.substring(0, headerEnd);
                var body = trimmed.substring(headerEnd + 5);
                try {
                    var h = jsyaml.safeLoad(header);
                    var r = "#" + h.title + "\n";
                    if (h.tags) {
                        r += "in: "
                        for (var i in h.tags) {
                            var tag = h.tags[i];
                            if (i > 0) r += ", ";
                            r += "[" + tag + "](#list/" + tag + ")"
                        }
                        r += "\n\n";
                    }
                    r += "* * *\n\n";
                    r += body;
                    r += "* * *\n\n";
                    if (h.authors) r += "editable by: " + h.authors.join(", ") + "\n\n";

                    source = r;
                } catch (e) {
                    console.log(e);
                    source = "<pre>\n" + header + "</pre>\n" + body;
                }
            }
        }

        return markdown.makeHtml(source);
    };
    return {
        restrict: 'E',
        replace: true,
        link: function (scope, element, attrs) {
            scope.$watch(attrs.ngModel, function(source) {
                element.html(convert(source));
            });
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
        who: function() {
            return api("who");
        },
        list: function(query) {
            return api("list", {query: query});
        },
        get: function(key) {
            return api("get", {key: key});
        },
        put: function(key, content) {
            return api("put", {key:key, content: content});
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
.value("serviceEndpoint", "http://localhost:8023")
;
