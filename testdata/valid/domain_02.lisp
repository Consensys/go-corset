(defcolumns (STAMP :i12))
;; STAMP[-1] == 0
(defconstraint c1 (:domain {-1}) (== 0 STAMP))
