(defpurefun (vanishes! x) (== 0 x))

(defcolumns (COUNTER :i10))
(defalias CT3 CT2)
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () (vanishes! CT3))
