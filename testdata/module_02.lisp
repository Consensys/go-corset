(defpurefun ((vanishes! :@loob) x) x)
(defcolumns X)
(module test)
(defcolumns X)
(defconstraint heartbeat () (vanishes! X))
