(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (X :i16))
(module test)
(defcolumns (X :i16))
(defconstraint heartbeat () (vanishes! X))
