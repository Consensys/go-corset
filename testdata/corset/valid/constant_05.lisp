(defconst ONE 1)
(defcolumns (CT :i4))
(defconstraint c1 ()
  (== 0
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
