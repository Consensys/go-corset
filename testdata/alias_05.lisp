(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns COUNTER)
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () (vanishes! CT2))
