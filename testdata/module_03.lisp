(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16))
(module m1)
;; Module without any column declarations to test alignment.
(module m2)
(defcolumns (X :i16))
(defconstraint heartbeat () (vanishes! X))
