(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (COUNTER :i32))
(defalias CT COUNTER)
(defconstraint heartbeat () (vanishes! CT))
