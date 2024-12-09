(defcolumns COUNTER)
(defalias CT1 COUNTER)
(defalias CT2 CT1)
(defconstraint heartbeat () CT2)
