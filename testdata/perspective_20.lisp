(defpurefun ((vanishes! :@loob :force) e0) e0)

(module m1)
(defcolumns A (P :binary@prove))
(defperspective p1 P ((B :binary)))
(defconstraint c1 (:perspective p1) (vanishes! (- A B)))

(module m2)
(defcolumns A (P :binary@prove))
(defperspective p2 P ((B :binary)))
(defconstraint c1 (:perspective p2) (vanishes! (* A B)))
