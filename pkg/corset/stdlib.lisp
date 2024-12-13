(defpurefun ((vanishes! :@loob) x) x)

;; =============================================================================
;; Logical
;; =============================================================================
(defpurefun ((or! :@loob) x y) (* x y))
(defpurefun ((eq! :@loob) x y) (- x y))

;; =============================================================================
;; Control Flow
;; =============================================================================
(defpurefun (if-not-eq lhs rhs then)
    (if
     (eq! lhs rhs)
     ;; True branch
     (vanishes! 0)
     ;; False branch
     then))

(defpurefun (if-eq lhs rhs then)
    (if
     (eq! lhs rhs)
     ;; True branch
     then))

(defpurefun (if-eq-else lhs rhs then else)
    (if
     (eq! lhs rhs)
     ;; True branch
     then
     ;; False branch
     else))

;; =============================================================================
;; Temporal
;; =============================================================================
(defpurefun (prev x) (shift x -1))
(defpurefun (next x) (shift x 1))
(defpurefun (will-eq! e0 e1) (eq! (next e0) e1))
(defpurefun (will-inc! e0 offset) (will-eq! e0 (+ e0 offset)))
(defpurefun (will-remain-constant! e0) (will-eq! e0 e0))
