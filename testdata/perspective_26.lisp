(defpurefun ((vanishes! :𝔽@loob :force) x) x)
;;
(defcolumns
    ;; Column (not in perspective)
    A
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

;; Section 1
(defperspective p1 P ((B :binary)))
(defun (p1-B) p1/B)
(defconstraint c1 (:perspective p1) (vanishes! (- A (p1-B))))

;; Section 2
(defperspective p2 Q ((C :binary)))
(defun (p2-C) p2/C)
(defconstraint c2 (:perspective p2) (vanishes! (* A (p2-C))))
