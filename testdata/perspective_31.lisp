(defcolumns
  (ST :i16@loob@prove)
  (P :binary@prove))
;;
(defperspective p1 P ((X :i16@loob@prove)))
(defpermutation (ST' Y) ((↓ ST) (↑ p1/X)))
(defconstraint first-row (:domain {-1}) (- Y 5))
