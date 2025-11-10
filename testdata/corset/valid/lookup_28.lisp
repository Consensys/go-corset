(defcolumns (X :i16) (Y :i16) (P :i16))
(defpurefun (prev (X :any)) (shift X -1))
;; define selector function
(defun ((selector :binary :force)) (- P (prev P)))
;; example use of selector
(defclookup l1 (Y) (selector) (X))

;; enforce (P - (prev P)) is binary.
(defconstraint inc ()
  (∨
   (== P (prev P))
   (== (- P 1) (prev P))))
