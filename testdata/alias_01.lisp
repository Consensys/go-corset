(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns COUNTER)
(defalias CT COUNTER)
(defconstraint heartbeat () (vanishes! CT))
