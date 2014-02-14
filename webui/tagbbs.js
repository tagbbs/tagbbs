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
    $scope.error = "";
    $scope.post = {};
    $scope.show_raw = function() {
        window.open('data:text/plain;charset=utf-8,' + encodeURIComponent($scope.post.content));
    };
    $scope.$watch("key", function(key) {
        bbs.get(key).success(function(d) {
            $scope.error = d.error;
            $scope.post = d.result || {};
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
            "title: \"Your Title\"\n" +
            "authors: [yourid]\n" +
            "tags: [test]\n" +
            "---\n\n" +
            "Content\n";
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
.directive("post", function () {
    var markdown = new Showdown.converter();
    var convert = function(post) {
        if (!post) return "";
        var source = post.content;
        if (!source) return "";
        var trimmed = source.trimLeft();
        if (trimmed.substring(0, 4) == "---\n") {
            var headerEnd = trimmed.indexOf("\n---\n");
            if (headerEnd > 0) {
                var header = trimmed.substring(0, headerEnd);
                var body = trimmed.substring(headerEnd + 5);
                try {
                    var h = jsyaml.safeLoad(header);
                    var r = ""
                    if (h.title) r += "#" + h.title + "\n\n";
                    r += "rev: " + post.rev + ", " + post.timestamp + "\n\n";
                    if (h.tags) {
                        r += "in: "
                        for (var i in h.tags) {
                            var tag = h.tags[i];
                            if (i > 0) r += ", ";
                            r += "[" + tag + "](#list/" + tag + ")"
                        }
                        r += "\n\n";
                    }
                    if (h.authors) {
                        r += "by: ";
                        h.authors.forEach(function(a) {
                            r += "[" + a + "](#show/user:" + a + ")"
                        });
                        r += "\n\n";
                    }
                    r += "* * *\n\n";
                    r += body + "\n\n";

                    source = r;
                } catch (e) {
                    console.log(e);
                    source = "rev: " + post.rev + ", " + post.timestamp + "\n\n" + "<pre>\n" + header + "\n---\n" + "</pre>\n" + body;
                }
            }
        }

        return markdown.makeHtml(source);
    };
    return {
        require: "ngModel",
        restrict: "E",
        replace: true,
        link: function (scope, element, attrs, ngModel) {
            scope.$watch(attrs.ngModel, function(post) {
                element.html(convert(post));
            });
        }
    };
})
.directive("codemirror", function() {
return {
        require: "ngModel",
        restrict: "A",
        link: function (scope, elm, attrs, ngModel) {
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
.value("serviceEndpoint", location.protocol + "//" + location.hostname + ":8023")
;
