(defpurefun (vanishes! x) (== 0 x))

(defcolumns (COUNTER :i16))
(defalias CT1 COUNTER)
(defalias CT2 CT1)
(defconstraint heartbeat () (vanishes! CT2))
