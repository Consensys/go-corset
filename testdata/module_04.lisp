(module m1)
(defcolumns X)
;; Module without any column declarations to test alignment.
(module m2)
(defcolumns X)
(defconstraint heartbeat () X)
