(defpurefun ((vanishes! :@loob) x) x)

(defcolumns COUNTER)
(defalias CT COUNTER)
(defconstraint heartbeat () (vanishes! CT))
