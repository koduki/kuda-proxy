package fw;

import java.io.IOException;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.io.UncheckedIOException;
import java.util.Map;
import javax.enterprise.context.ApplicationScoped;
import javax.ws.rs.core.Response;
import javax.ws.rs.ext.ExceptionMapper;
import javax.ws.rs.ext.Provider;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import dev.nklab.jl2.web.logging.Logger;

@ApplicationScoped
@Provider
public class JsonExceptionMapper implements ExceptionMapper<Exception> {

    private static final Logger LOGGER = Logger.getLogger("slide4vr");

    @Override
    public Response toResponse(Exception exception) {
        var errorMessage = (exception.getMessage() == null) ? "" : exception.getMessage();
        var msg = Map.of("error", exception.getClass().getName(), "message", errorMessage);

        return toInternalErrorResponse(exception, msg);
    }

    Response toInternalErrorResponse(Exception exception, Map<String, String> msg) {
        try {
            LOGGER.severe(exception);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                    .entity(new ObjectMapper().writeValueAsString(msg)).build();
        } catch (JsonProcessingException ex) {
            throw new RuntimeException(ex);
        }
    }

    String parseStackTrace(Exception exception) {
        try ( var sw = new StringWriter();  var pw = new PrintWriter(sw);) {
            exception.printStackTrace(pw);
            pw.flush();

            return sw.toString();
        } catch (IOException ex) {
            throw new UncheckedIOException(ex);
        }
    }
}
