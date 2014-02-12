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
.controller("MainCtrl", function($scope, $http) {
})
.controller("Login", function($scope, $location, bbs) {
    $scope.user = "";
    $scope.pass = "";
    $scope.message = "";
    $scope.submit = function() {
        bbs.login($scope.user, $scope.pass).success(function(d) {
            if (d.result) {
                $location.path("/list");
            } else {
                $scope.message = d.error;
            }
        });
    }
})
.controller("Logout", function($scope, $location, bbs) {
    $scope.error = "logging out...";
    bbs.logout().success(function(d) {
        if (d.error) {
            $scope.error = d.error;
        } else {
            $location.path('/login');
        }
    });
})
.controller("Register", function($scope) {

})
.controller("List", function($scope, $routeParams, bbs) {
    $scope.query = "";
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
                headerEnd = headerEnd + 5;
                var header = trimmed.substring(0, headerEnd);
                var body = trimmed.substring(headerEnd);
                source = "<pre>\n" + header + "</pre>\n" + body;
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
