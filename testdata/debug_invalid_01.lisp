(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- (debug X) Y)))
