(defcolumns (X :i16) (Y :i16) (P :i16))

(defun (selector) (force-bin (- P (prev P))))

;; example use of selector
(defclookup l1 (Y) (selector) (X))

;; enforce (P - (prev P)) is binary.
(defconstraint inc ()
  (∨
   (== P (prev P))
   (== (- P 1) (prev P))))
