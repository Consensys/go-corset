(defpurefun (vanishes! x) (== 0 x))

(defcolumns (COUNTER :i12))
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () (vanishes! CT2))
