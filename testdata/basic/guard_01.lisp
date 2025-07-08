(defpurefun (vanishes! x) (== 0 x))

(defcolumns (STAMP :i4) (X :i16))
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard STAMP)
  (vanishes! (* (- X 1) (- X 2))))
