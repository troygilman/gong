package gong

// func buildRealPath(route Route, request *http.Request) string {
// 	routePathSplit := strings.Split(route.FullPath(), "/")
// 	requestPathSplit := strings.Split(request.URL.EscapedPath(), "/")
// 	for i, routePathFragment := range routePathSplit {
// 		if i >= len(requestPathSplit) {
// 			continue
// 		}
// 		requestPathFragment := requestPathSplit[i]
// 		if routePathFragment == requestPathFragment {
// 			continue
// 		}
// 		if strings.HasPrefix(routePathFragment, "{") && strings.HasSuffix(routePathFragment, "}") {
// 			routePathSplit[i] = requestPathFragment
// 		}
// 	}
// 	return strings.Join(routePathSplit, "/")
// }
