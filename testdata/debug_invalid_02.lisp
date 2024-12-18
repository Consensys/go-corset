(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- X (* (debug X) Y))))
