(defcolumns (X :i16) (Y :i16))
(defconst FORK 3)

(defconstraint c1 ()
  (if (< FORK 5)
      (== X (+ Y Y))))

(defconstraint c2 ()
  (if (> FORK 5)
      (== X Y)))
