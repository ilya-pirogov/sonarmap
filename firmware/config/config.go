package config

const (
    NetCatCmd = "nohup nc -l -p %s > %s &"
    VerFile = "/etc/sonarmap"
    DstSonarMap = "/usr/sbin/sonarmap"
    TranslationsDir="/usr/share/NOS/translations"
    Username = "root"
    Password = "nos"
    RcPatch = `--- /etc/rc.d/rc.S.orig
+++ /etc/rc.d/rc.S
@@ -627,6 +627,10 @@
     fi
 fi

+if [ -x /usr/sbin/sonarmap ]; then
+    /usr/sbin/sonarmap &
+fi
+
 # Background waiting, so we don't prevent login prompt
 (
   wait_pid "${RC_MODULES_PID:-}" "modules to load"        10`
)
//
//const (
//    NetCatCmd = "nc -l -p %s > %s &"
//    VerFile = "/home/ilya/sonarmap.ver"
//    DstSonarMap = "/home/ilya/sonarmap"
//    Username = "ilya"
//    Password = "bk39mmz"
//    RcPatch = `--- /home/ilya/rc.S.orig
//+++ /home/ilya/rc.S
//@@ -627,6 +627,10 @@
//     fi
// fi
//
//+if [ -x /home/ilya/sonarmap ]; then
//+    /home/ilya/sonarmap &
//+fi
//+
// # Background waiting, so we don't prevent login prompt
// (
//   wait_pid "${RC_MODULES_PID:-}" "modules to load"        10`
//)
//
