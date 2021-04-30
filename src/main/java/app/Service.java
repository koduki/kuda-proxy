/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package app;

import com.google.cloud.datastore.Datastore;
import com.google.cloud.datastore.DatastoreOptions;

/**
 *
 * @author koduki
 */
public class Service {

    public static Datastore datastore() {
        var options = DatastoreOptions.getDefaultInstance();
        var datastore = options.getService();
        return datastore;
    }

}
