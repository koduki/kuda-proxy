package app;

import static dev.nklab.jl2.Extentions.$;
import dev.nklab.jl2.web.logging.Logger;
import dev.nklab.kuda.core.Trigger;
import dev.nklab.jl2.web.profile.DistributedTracer;
import io.opencensus.contrib.http.util.HttpTraceAttributeConstants;
import io.opencensus.trace.AttributeValue;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Map;
import java.util.stream.Collectors;
import javax.inject.Inject;
import javax.servlet.http.HttpServletRequest;
import javax.ws.rs.Consumes;
import javax.ws.rs.POST;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@Path("/")
public class ProxyResource {

    private final Logger logger = Logger.getLogger("kuda-proxy");

    @Inject
    Trigger trigger;

    @ConfigProperty(name = "kuda.proxy.target.name")
    String targetName;

    @ConfigProperty(name = "kuda.proxy.target.url")
    String targetUrl;

    @ConfigProperty(name = "kuda.proxy.target.method")
    String targetMethod;

    @ConfigProperty(name = "kuda.proxy.trigger.condkey")
    String condKey;

    @ConfigProperty(name = "dev.nklab.profile.trace")
    boolean isTrace;

    @Inject
    HttpServletRequest forwardRequest;

    @POST
    @Consumes(MediaType.APPLICATION_JSON)
    @Produces(MediaType.APPLICATION_JSON)
    public String forward(Map<String, Object> params) {
        return DistributedTracer.trace().isTrace(isTrace).apply("kuda-proxy", forwardRequest, (t) -> {
            try {
                // Parse request
                System.out.println(params.get("key"));
                var targetParams = (Map<String, String>) params.get("targetParams");

                var query = "";
                if ("GET".equals(targetMethod)) {
                    query = String.join("&",
                            targetParams.entrySet().stream()
                                    .map(x -> x.getKey() + "=" + x.getValue()).collect(Collectors.toList())
                    );
                }

                // Exec target application
                logger.debug("target-start", $("", ""));
                callTarget(forwardRequest, targetMethod, query);

                logger.debug("trigger-start", $("condKey", condKey));
                return trigger.callTrigger(params, condKey);
            } catch (IOException | InterruptedException ex) {
                throw new RuntimeException(ex);
            }
        });
    }

    HttpResponse callTarget(HttpServletRequest forwardRequest, String targetMethod, String query) throws InterruptedException, IOException {
        var path = "/?" + query;
        var url = targetUrl + path;

        return DistributedTracer.trace().isTrace(isTrace).apply(targetName, forwardRequest,
                "",
                "GET",
                path,
                (t) -> {
                    try {
                        var client = HttpClient.newHttpClient();
                        var req = HttpRequest.newBuilder()
                                .GET()
                                .version(HttpClient.Version.HTTP_1_1)
                                .header("Content-Type", "application/json")
                                .uri(URI.create(url))
                                .build();

                        var res = client.send(req, HttpResponse.BodyHandlers.ofString());
                        if (t.isPresent()) {
                            var span = t.get();
                            span.putAttribute(HttpTraceAttributeConstants.HTTP_URL, AttributeValue.stringAttributeValue(url));
                            span.putAttribute(HttpTraceAttributeConstants.HTTP_STATUS_CODE, AttributeValue.longAttributeValue(res.statusCode()));
                        }
                        System.out.println("http_code:" + res.statusCode() + ", body:" + res.body());
                        return res;
                    } catch (IOException | InterruptedException ex) {
                        throw new RuntimeException(ex);
                    }
                });
    }
}
