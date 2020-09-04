/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package fw;

import static dev.nklab.jl2.Extentions.*;
import dev.nklab.jl2.web.logging.Logger;
import dev.nklab.jl2.web.profile.TracingBootstrap;
import java.io.IOException;
import javax.enterprise.context.ApplicationScoped;
import javax.enterprise.context.Initialized;
import javax.enterprise.event.Observes;
import javax.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;

/**
 *
 * @author koduki
 */
@ApplicationScoped
public class Bootstrap {

    @ConfigProperty(name = "kuda.proxy.target.name")
    String targetName;

    private final Logger logger = Logger.getLogger("slide4vr");

    @Inject
    TracingBootstrap tracingBootstrap;

    public void handle(@Observes @Initialized(ApplicationScoped.class) Object event) throws IOException {
        tracingBootstrap.init();
        logger.info("init", $("target application", targetName));
    }
}
